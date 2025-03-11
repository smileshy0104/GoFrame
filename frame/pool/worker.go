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
// 该方法负责将worker添加到运行状态，并启动一个新的协程来执行具体任务。
func (w *Worker) run() {
	// 增加正在运行的worker
	w.pool.incRunning()
	// 使用协程运行worker
	go w.running()
}

// running worker运行
// 该方法是worker的主执行逻辑，负责从任务队列中取出任务并执行。
// 它还处理了任务执行过程中的异常情况，并确保worker在任务执行完毕后能够被正确回收和唤醒等待的worker。
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
