package sync

import (
	"runtime"
	"sync"
	"sync/atomic"
)

//spinlock本质就是一个unit32的正整数
type spinLock uint32

const maxBackoff = 16

func (sl *spinLock) Lock() {
	backoff := 1
	for !atomic.CompareAndSwapUint32((*uint32)(sl), 0, 1) {
		for i := 0; i < backoff; i++ {
			runtime.Gosched()
		}

		if backoff < maxBackoff {
			backoff <<= 1 //左移乘以2   右移除以2
		}
	}
}

func (sl *spinLock) Unlock() {
	atomic.StoreUint32((*uint32)(sl), 0)
}

func NewSpinLock() sync.Locker {
	return new(spinLock)
}
