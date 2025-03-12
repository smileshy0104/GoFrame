package pool

import (
	"math"
	"runtime"
	"sync"
	"testing"
	"time"
)

// 定义二进制前缀常量，用于内存大小的测量和转换
const (
	_   = 1 << (10 * iota) // 1 KiB = 1024 bytes
	KiB                    // 1 KiB = 1024 bytes
	MiB                    // 1 MiB = 1048576 bytes
	// GiB // 1 GiB = 1073741824 bytes
	// TiB // 1 TiB = 1099511627776 bytes (超过了int32的范围)
	// PiB // 1 PiB = 1125899906842624 bytes
	// EiB // 1 EiB = 1152921504606846976 bytes
	// ZiB // 1 ZiB = 1180591620717411303424 bytes (超过了int64的范围)
	// YiB // 1 YiB = 1208925819614629174706176 bytes
)

// 定义测试相关的常量
const (
	Param    = 100     // 测试参数
	PoolSize = 1000    // 线程池大小
	TestSize = 10000   // 测试任务数量
	n        = 1000000 // 循环次数
)

// curMem 用于记录当前分配的内存量（单位：字节）
var curMem uint64

// 定义运行参数和默认的过期时间常量
const (
	RunTimes           = 1000000          // 运行次数
	BenchParam         = 10               // 基准测试参数，用于模拟任务执行的时间
	DefaultExpiredTime = 10 * time.Second // 默认的过期时间
)

// demoFunc 是一个示例函数，用于模拟任务执行，每次调用会休眠一段时间
func demoFunc() {
	time.Sleep(time.Duration(BenchParam) * time.Millisecond)
}

// TestNoPool 测试不使用线程池的情况下的内存使用情况
//
// 参数:
// - t: 测试框架提供的测试对象，用于记录测试结果和日志输出
//
// 该测试函数通过启动多个 goroutine 来并发执行 `demoFunc` 函数，并在所有任务完成后统计内存使用情况。
// 具体步骤如下：
// 1. 创建一个 WaitGroup 用于等待所有 goroutine 完成。
// 2. 启动 n 个 goroutine，每个 goroutine 执行 `demoFunc` 并在完成后通知 WaitGroup。
// 3. 等待所有 goroutine 完成后，读取并记录内存使用情况。
func TestNoPool(t *testing.T) {
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			demoFunc()
			wg.Done()
		}()
	}

	wg.Wait()

	// 获取当前内存使用情况
	mem := runtime.MemStats{}
	runtime.ReadMemStats(&mem)
	curMem = mem.TotalAlloc/MiB - curMem
	t.Logf("memory usage:%d MB", curMem)
}

// TestHasPool 测试使用线程池的情况下的内存使用情况
//
// 参数:
// - t: 测试框架提供的测试对象，用于记录测试结果和日志输出
//
// 该测试函数通过创建一个线程池来管理并发任务的执行，并在所有任务完成后统计内存使用情况。
// 具体步骤如下：
// 1. 创建一个最大容量为 math.MaxInt32 的线程池，并确保在测试结束时释放资源。
// 2. 创建一个 WaitGroup 用于等待所有任务完成。
// 3. 提交 n 个任务到线程池中，每个任务执行 `demoFunc` 并在完成后通知 WaitGroup。
// 4. 等待所有任务完成后，读取并记录内存使用情况。
// 5. 记录当前正在运行的工作线程数和空闲工作线程数。
func TestHasPool(t *testing.T) {
	pool, _ := NewPool(math.MaxInt32) // 创建一个最大容量为 math.MaxInt32 的线程池
	defer pool.Release()              // 确保在测试结束时释放线程池资源
	var wg sync.WaitGroup

	for i := 0; i < n; i++ {
		wg.Add(1)
		_ = pool.Submit(func() {
			demoFunc()
			wg.Done()
		})
	}
	wg.Wait()

	// 获取当前内存使用情况
	mem := runtime.MemStats{}
	runtime.ReadMemStats(&mem)
	curMem = mem.TotalAlloc/MiB - curMem
	t.Logf("memory usage:%d MB", curMem)

	// 记录线程池的状态信息
	t.Logf("running worker:%d", pool.Running())
	t.Logf("free worker:%d ", pool.Free())
}
