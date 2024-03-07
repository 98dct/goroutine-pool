package goroutine_pool

import (
	"errors"
	syncx "goroutine-pool/tool/sync"
	"goroutine-pool/tool/wait"
	"sync"
	"sync/atomic"
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
	victim sync.Pool //sync.Pool用于存储和复用临时对象，并发安全
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

	p.victim.New = func() interface{} {
		return newGoWorker(&p)
	}
	p.cond = sync.NewCond(p.lock)

	go p.goPurge()

	return &p
}

func (p *Pool) Waiting() int32 {
	return atomic.LoadInt32(&p.waiting)
}

func (p *Pool) Running() int32 {
	return atomic.LoadInt32(&p.running)
}

func (p *Pool) Cap() int32 {
	return p.capacity
}

func (p *Pool) Free() int32 {
	if p.Cap() < 0 {
		return -1
	}
	return p.Cap() - p.Running()
}

//调整容量
func (p *Pool) Tune(capacity int32) {
	oldCap := p.Cap()
	if capacity < 0 {
		capacity = -1
	}
	if oldCap == capacity {
		return
	}

	atomic.StoreInt32(&p.capacity, capacity)

	if p.Cap() < 0 {
		p.cond.Broadcast()
		return
	}

	if oldCap > 0 && p.Cap() > oldCap {
		p.cond.Broadcast()
	}
}

//外部提交任务
func (p *Pool) Submit(task TaskFunc) error {
	if p.IsClosed() {
		return ErrPoolCLosed
	}

	w, err := p.getGoWorker()
	if err != nil {
		return err
	}
	w.exec(task)
	return nil
}

func (p *Pool) getGoWorker() (*goWorker, error) {

	//上锁，避免并发问题
	p.allRunning.Add(1)
	p.lock.Lock()
	defer func() {
		p.lock.Unlock()
		p.allRunning.Add(-1)
	}()

	for {
		if p.IsClosed() {
			return nil, ErrPoolClosed
		}

		//1.从local缓冲中获取
		if w := p.local.detach(); w != nil {
			return w, nil
		}

		//2.从victim中获取（新建一个）
		if p.Cap() == -1 || p.Cap() > atomic.LoadInt32(&p.running) {
			//还有容量
			raw := p.victim.Get()
			w, ok := raw.(*goWorker)
			if !ok {
				return nil, errors.New("victim cache data is wrong type")
			}
			w.start()
			return w, nil
		}

		//3.执行到这里，说明没有空闲的协程对象
		//非阻塞模式 or 阻塞模式（但是阻塞的太多了）
		if p.option.NonBlocking || (p.option.MaxBlockingTasks != 0 && p.Waiting() >= p.option.MaxBlockingTasks) {
			return nil, ErrPoolOverload
		}

		//4.阻塞等待
		atomic.AddInt32(&p.waiting, 1)
		p.cond.Wait() //这里会对p.lock解锁，然后阻塞等待；被喊醒后，又对p.lock上锁
		atomic.AddInt32(&p.waiting, -1)
	}

}

func (p *Pool) Close() {
	close(p.closed)

	p.lock.Lock()
	p.local.reset()
	p.lock.Unlock()
	p.cond.Broadcast()

	p.allRunning.Wait()
}

func (p *Pool) CloseWithTimeOut(timeout time.Duration) error {
	close(p.closed)

	p.lock.Lock()
	p.local.reset()
	p.lock.Unlock()
	p.cond.Broadcast()

	p.allRunning.WaitWithTimeOut(timeout)

	if p.Running() != 0 || p.Waiting() != 0 || (!p.option.DisablePurge && atomic.LoadInt32(&p.purgeDone) == 0) {
		return ErrTimeout
	}
	return nil
}

func (p *Pool) IsClosed() bool {
	select {
	case <-p.closed:
		return true
	default:

	}

	return false
}
