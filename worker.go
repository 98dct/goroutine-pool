package goroutine_pool

import (
	"sync/atomic"
	"time"
)

//协程对象用于处理任务
type TaskFunc func()

type goWorker struct {
	tashChan chan TaskFunc
	pool     *Pool

	//最后一次使用的时间
	lastUsedTime time.Time
}

func newGoWorker(pool *Pool) *goWorker {

	gw := goWorker{}
	gw.pool = pool
	gw.tashChan = make(chan TaskFunc)
	return &gw
}

//启动协程（1、等待任务  2、任务处理完成 自动会收到local中）
func (gw *goWorker) start(){

	atomic.AddInt32(&gw.pool.running,1)
	gw.pool.allRunning.Add(1)
	go func() {
		defer func() {
			atomic.AddInt32(&gw.pool.running,-1)
			gw.pool.allRunning.Add(-1)
			gw.pool.victim.Put(gw)  //执行失败，回收到victim中

			if p := recover();p != nil{
				handler := gw.pool.
			}
		}()
	}()
}
