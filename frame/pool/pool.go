package pool

import (
	"errors"
	"fmt"
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
	//cond
	cond *sync.Cond
	//PanicHandler
	PanicHandler func()
}

// NewPool 创建一个Pool
func NewPool(cap int) (*Pool, error) {
	// 调用NewTimePool创建一个默认到期时间的Pool
	return NewTimePool(cap, DefaultExpire)
}

//func NewPoolConf() (*Pool, error) {
//	cap, ok := config.Conf.Pool["cap"]
//	if !ok {
//		return nil, errors.New("cap config not exist")
//	}
//	return NewTimePool(int(cap.(int64)), DefaultExpire)
//}

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

func (p *Pool) expireWorker() {
	//定时清理过期的空闲worker
	ticker := time.NewTicker(p.expire)
	for range ticker.C {
		if p.IsClosed() {
			break
		}
		//循环空闲的workers 如果当前时间和worker的最后运行任务的时间 差值大于expire 进行清理
		p.lock.Lock()
		idleWorkers := p.workers
		n := len(idleWorkers) - 1
		if n >= 0 {
			var clearN = -1
			for i, w := range idleWorkers {
				if time.Now().Sub(w.lastTime) <= p.expire {
					break
				}
				clearN = i
				w.task <- nil
				idleWorkers[i] = nil
			}
			// 3 2
			if clearN != -1 {
				if clearN >= len(idleWorkers)-1 {
					p.workers = idleWorkers[:0]
				} else {
					// len=3 0,1 del 2
					p.workers = idleWorkers[clearN+1:]
				}
				fmt.Printf("清除完成,running:%d, workers:%v \n", p.running, p.workers)
			}
		}
		p.lock.Unlock()
	}
}

func (p *Pool) waitIdleWorker() *Worker {
	p.lock.Lock()
	p.cond.Wait()

	idleWorkers := p.workers
	n := len(idleWorkers) - 1
	if n < 0 {
		p.lock.Unlock()
		if p.running < p.cap {
			//还不够pool的容量，直接新建一个
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
			w.run()
			return w
		}
		return p.waitIdleWorker()
	}
	w := idleWorkers[n]
	idleWorkers[n] = nil
	p.workers = idleWorkers[:n]
	p.lock.Unlock()
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
func (p *Pool) PutWorker(w *Worker) {
	w.lastTime = time.Now()
	p.lock.Lock()
	p.workers = append(p.workers, w)
	p.cond.Signal()
	p.lock.Unlock()
}

func (p *Pool) Release() {
	p.once.Do(func() {
		//只执行一次
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
		p.lock.Unlock()
		p.release <- sig{}
	})
}

func (p *Pool) IsClosed() bool {

	return len(p.release) > 0
}

func (p *Pool) Restart() bool {
	if len(p.release) <= 0 {
		return true
	}
	_ = <-p.release
	go p.expireWorker()
	return true
}

func (p *Pool) Running() int {
	return int(atomic.LoadInt32(&p.running))
}

func (p *Pool) Free() int {
	return int(p.cap - p.running)
}
