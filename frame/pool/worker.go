package pool

import (
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
