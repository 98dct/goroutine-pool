package goroutine_pool

import (
	syncx "goroutine-pool/tool/sync"
	"goroutine-pool/tool/wait"
	"sync"
	"time"
)

const (
	maxIdleTime = 1 * time.Second
)

// goroutine池，用来管理workers
type Pool struct {

	//协程对象的最大值
	capacity int32
	//活跃的协程对象   atomic.AddIn32计算
	running int32

	//阻塞模式
	waiting int32      // 阻塞等待的数量
	cond    *sync.Cond //

	//pool是否关闭
	closed chan struct{}
	//配置信息
	option *Options

	//回收的对象
	victim sync.Pool
	//操作local对象的锁
	lock sync.Locker
	//协程对象缓存
	local goWorkerCache

	// 判断goPurge是否结束
	purgeDone int32
	// 等待协程对象停止 (包括协程池 和用户协程) 对waitgroup的封装
	allRunning *wait.Wait
}

func NewPool(capacity int32, opts ...Option) *Pool {
	//TODO 既然使用了选项模式，为什么不加一个default Options的配置呢？ 可以加个默认的option
	option := new(Options)
	for _, opt := range opts {
		opt(option)
	}

	//如果启用清理，清理间隔的值需要大于0
	if !option.DisablePurge {
		if option.MaxIdleTime <= 0 {
			option.MaxIdleTime = maxIdleTime
		}
	}

	//日志打印器
	if option.Logger == nil {
		option.Logger = defaultLogger
	}

	p := Pool{
		capacity:   capacity,
		closed:     make(chan struct{}),
		option:     option,
		lock:       syncx.NewSpinLock(),
		local:      NewCacheStack(),
		allRunning: &wait.Wait{},
	}

	return &p
}
