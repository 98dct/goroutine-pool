package goroutine_pool

import "time"

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
