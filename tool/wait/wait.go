package wait

import (
	"sync"
	"time"
)

// 对sync.waitgroup进行封装
type Wait struct {
	wait sync.WaitGroup
}

func (w *Wait) Add(num int) {
	w.wait.Add(num)
}

func (w *Wait) Done() {
	w.wait.Done()
}

func (w *Wait) Wait() {
	w.wait.Wait()
}

//超时等待
//设置一组goroutine的执行时间，超过某个时间，就返回true 即超时了
func (w *Wait) WaitWithTimeOut(timeout time.Duration) bool {

	ch := make(chan struct{})
	go func() {
		defer close(ch)
		w.Wait()
	}()

	select {
	case <-ch:
		return false //正常
	case <-time.After(timeout):
		return true // 超时
	}

}
