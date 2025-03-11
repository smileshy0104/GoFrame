package pool

import (
	"sync"
	"time"
)

// sig 结构体 存放一个信号
type sig struct{}

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
