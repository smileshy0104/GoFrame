package pool

import (
	"errors"
	"fmt"
	"frame/config"
	"sync"
	"sync/atomic"
	"time"
)

// sig 结构体 存放一个信号
type sig struct{}

// DefaultExpire 默认过期时间
const DefaultExpire = 5

// 自定义错误
var (
	ErrorInValidCap    = errors.New("pool cap can not <= 0")
	ErrorInValidExpire = errors.New("pool expire can not <= 0")
	ErrorHasClosed     = errors.New("pool has bean released")
)

// Pool 结构体
type Pool struct {
	//cap 容量 pool max cap
	cap int32
	//running 正在运行的worker的数量
	running int32
	//空闲worker(相当于可以处理的Task的processor)
	workers []*Worker
	//expire 过期时间 空闲的worker超过这个时间 回收掉
	expire time.Duration
	//release 释放资源  pool就不能使用了
	release chan sig
	//lock 去保护pool里面的相关资源的安全
	lock sync.Mutex
	//once 释放只能调用一次 不能多次调用
	once sync.Once
	//workerCache 缓存
	workerCache sync.Pool
	//cond 用于条件变量同步
	cond *sync.Cond
	//PanicHandler 用于处理 panic 情况
	PanicHandler func()
}

// NewPool 创建一个Pool
func NewPool(cap int) (*Pool, error) {
	// 调用NewTimePool创建一个默认到期时间的Pool
	return NewTimePool(cap, DefaultExpire)
}

// NewPoolConf 根据配置文件创建一个Pool
func NewPoolConf() (*Pool, error) {
	// 获取配置文件内容
	cap, ok := config.Conf.Pool["cap"]
	// 如果没有配置文件，返回错误
	if !ok {
		return nil, errors.New("cap config not exist")
	}
	// 使用类型断言来获取配置文件中的cap，再调用NewTimePool创建一个默认到期时间的Pool
	return NewTimePool(int(cap.(int64)), DefaultExpire)
}

// NewTimePool 创建一个Pool
func NewTimePool(cap int, expire int) (*Pool, error) {
	// 容量不能小于等于0
	if cap <= 0 {
		return nil, ErrorInValidCap
	}
	// 过期时间不能小于等于0
	if expire <= 0 {
		return nil, ErrorInValidExpire
	}
	// 实例化Pool
	p := &Pool{
		cap:     int32(cap),
		expire:  time.Duration(expire) * time.Second,
		release: make(chan sig, 1),
	}

	return p, nil
}

// Submit 提交一个任务
func (p *Pool) Submit(task func()) error {
	// 判断是否pool资源已经被释放
	if len(p.release) > 0 {
		return ErrorHasClosed
	}

	//获取池里面的一个worker，然后执行任务就可以了
	w := p.GetWorker()
	w.task <- task
	return nil
}

// GetWorker 获取指定pool中的worker
func (p *Pool) GetWorker() (w *Worker) {
	//1. 目的获取pool里面的worker
	readyWorker := func() {
		w = p.workerCache.Get().(*Worker)
		w.run()
	}
	p.lock.Lock()
	idleWorkers := p.workers
	//2. 如果 有空闲的worker 直接获取
	n := len(idleWorkers) - 1
	if n >= 0 {
		w = idleWorkers[n]
		idleWorkers[n] = nil
		p.workers = idleWorkers[:n]
		p.lock.Unlock()
		return
	}
	//3. 如果没有空闲的worker，要新建一个worker（对应pool具有空闲）
	if p.running < p.cap {
		p.lock.Unlock()
		//还不够pool的容量，直接新建一个
		//c := p.workerCache.Get()
		//var w *Worker
		//if c == nil {
		//	w = &Worker{
		//		pool: p,
		//		task: make(chan func(), 1),
		//	}
		//} else {
		//	w = c.(*Worker)
		//}
		//w.run()
		readyWorker()
		return
	}
	p.lock.Unlock()
	//4. 如果正在运行的workers 如果大于pool容量，阻塞等待，worker释放
	//for {
	return p.waitIdleWorker()
	//}
}

// expireWorker 定期检查并清理过期的空闲worker。
//
// 该函数在一个独立的goroutine中运行，通过ticker按照设定的时间间隔进行清理操作。
// 每次清理时，会遍历所有空闲的worker，并将超过设定过期时间的worker从池中移除。
func (p *Pool) expireWorker() {
	// 创建一个定时器，每隔p.expire时间触发一次清理操作
	ticker := time.NewTicker(p.expire)

	// 持续监听ticker的事件，每次触发时执行一次清理操作
	for range ticker.C {
		// 如果池已关闭，则停止清理操作
		if p.IsClosed() {
			break
		}

		// 加锁以确保线程安全地访问和修改共享资源
		p.lock.Lock()

		// 获取当前所有的空闲worker
		idleWorkers := p.workers

		// 如果有空闲worker，则进行清理操作
		n := len(idleWorkers) - 1
		if n >= 0 {
			var clearN = -1

			// 遍历所有空闲worker，检查它们是否已经过期
			for i, w := range idleWorkers {
				// 如果当前worker未过期，则停止清理操作
				if time.Now().Sub(w.lastTime) <= p.expire {
					break
				}

				// 记录需要清理的最后一个worker的索引
				clearN = i

				// 向worker发送nil任务，使其退出运行
				w.task <- nil

				// 将worker置为nil，表示它已被清理
				idleWorkers[i] = nil
			}

			// 如果有worker被清理，则更新workers列表
			if clearN != -1 {
				if clearN >= len(idleWorkers)-1 {
					// 如果清理的是最后一个worker，则清空workers列表
					p.workers = idleWorkers[:0]
				} else {
					// 否则，截取清理后的workers列表
					p.workers = idleWorkers[clearN+1:]
				}

				// 打印清理完成的日志信息
				fmt.Printf("清除完成,running:%d, workers:%v \n", p.running, p.workers)
			}
		}

		// 解锁以释放共享资源
		p.lock.Unlock()
	}
}

// waitIdleWorker 等待并获取一个空闲的worker。
//
// 该函数首先尝试从现有的空闲worker中获取一个，如果没有任何空闲worker且pool未满员，
// 则创建一个新的worker；如果pool已满员，则等待直到有空闲的worker可用。
func (p *Pool) waitIdleWorker() *Worker {
	// 加锁以确保线程安全地访问和修改共享资源
	p.lock.Lock()

	// 等待条件变量，直到有空闲worker或池已关闭
	p.cond.Wait()

	// 获取当前所有的空闲worker
	idleWorkers := p.workers

	// 如果没有空闲worker，则根据pool的状态决定如何处理
	n := len(idleWorkers) - 1
	if n < 0 {
		p.lock.Unlock()

		// 如果pool未满员，则创建一个新的worker
		if p.running < p.cap {
			c := p.workerCache.Get()
			var w *Worker
			if c == nil {
				w = &Worker{
					pool: p,
					task: make(chan func(), 1),
				}
			} else {
				w = c.(*Worker)
			}

			// 启动新的worker
			w.run()
			return w
		}

		// 如果pool已满员，则继续等待空闲worker
		return p.waitIdleWorker()
	}

	// 从空闲worker列表中取出最后一个worker
	w := idleWorkers[n]
	idleWorkers[n] = nil

	// 更新workers列表，移除已取出的worker
	p.workers = idleWorkers[:n]

	// 解锁以释放共享资源
	p.lock.Unlock()

	// 返回获取到的worker
	return w
}

// incRunning 增加正在运行的worker数量
func (p *Pool) incRunning() {
	atomic.AddInt32(&p.running, 1)
}

// decRunning 减少正在运行的worker数量
func (p *Pool) decRunning() {
	atomic.AddInt32(&p.running, -1)
}

// Running 获取当前正在运行的工作协程数量。
func (p *Pool) Running() int {
	return int(atomic.LoadInt32(&p.running))
}

// PutWorker 将一个已存在的Worker放回Pool中，以便再次使用
// 此函数旨在回收完成任务的Worker，使其能够进入等待状态，并在需要时被重新调度
func (p *Pool) PutWorker(w *Worker) {
	// 记录一下worker的最后运行时间
	w.lastTime = time.Now()
	// 加锁以保护共享资源
	p.lock.Lock()
	// 放回worker
	p.workers = append(p.workers, w)
	// 唤醒一个等待的worker
	p.cond.Signal()
	// 解锁以释放共享资源
	p.lock.Unlock()
}

// Release 释放池资源，确保释放操作仅执行一次。该方法会等待所有工作协程空闲后才进行资源释放。
//
// 此函数通过 once.Do 确保只执行一次，并且在释放资源时会清空所有工作协程的任务和引用，最后发送一个信号表示资源已释放。
func (p *Pool) Release() {
	p.once.Do(func() {
		// 获取锁以确保线程安全地释放资源
		p.lock.Lock()
		workers := p.workers
		for i, w := range workers {
			if w == nil {
				continue
			}
			w.task = nil
			w.pool = nil
			workers[i] = nil
		}
		p.workers = nil
		// 释放锁
		p.lock.Unlock()
		// 发送信号表示资源已释放
		p.release <- sig{}
	})
}

// IsClosed 检查池是否已关闭。
func (p *Pool) IsClosed() bool {
	return len(p.release) > 0
}

// Restart 尝试重启池。如果池未关闭则直接返回 true；如果池已关闭，则接收释放信号并重新启动过期工作协程。
func (p *Pool) Restart() bool {
	if len(p.release) <= 0 {
		return true
	}
	_ = <-p.release
	go p.expireWorker()
	return true
}

// Free 获取池中空闲的可用工作协程数量。
func (p *Pool) Free() int {
	return int(p.cap - p.running)
}
