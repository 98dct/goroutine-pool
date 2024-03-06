package goroutine_pool

import "time"

type goWorkerCache interface {
	len() int                                       //协程对象个数
	isEmpty() bool                                  // 缓冲池是否为空
	insert(worker *goWorker) error                  //缓冲一个协程对象
	detach() *goWorker                              //获取一个协程对象
	reset()                                         //将缓冲中的协程对象全部清除
	clearExpire(duration time.Duration) []*goWorker //长时间未使用的协程对象

}
