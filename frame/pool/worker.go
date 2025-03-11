package pool

import (
	newlogger "frame/log"
	"time"
)

// Worker 结构体
type Worker struct {
	//持有的pool 池
	pool *Pool
	//task 任务队列（执行的任务）
	task chan func()
	//lastTime 执行任务的最后的时间
	lastTime time.Time
}

// run 运行worker
func (w *Worker) run() {
	// 增加正在运行的worker
	w.pool.incRunning()
	// 使用协程运行worker
	go w.running()
}

// running worker运行
func (w *Worker) running() {
	defer func() {
		// 减少正在运行的worker（相当于processor）
		w.pool.decRunning()
		// 将worker放回workerCache中（相当于processor）
		w.pool.workerCache.Put(w)
		// 捕获任务发生的panic
		if err := recover(); err != nil {
			//捕获任务发生的panic
			if w.pool.PanicHandler != nil {
				w.pool.PanicHandler()
			} else {
				newlogger.Default().Error(err)
			}
		}
		// 唤醒等待的worker
		w.pool.cond.Signal()
	}()
	// 循环执行任务task
	for f := range w.task {
		// 如果f为空，说明worker已经关闭了
		if f == nil {
			return
		}
		// 执行任务
		f()
		//任务运行完成，worker空闲
		w.pool.PutWorker(w)
	}
}
