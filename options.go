package goroutine_pool

import (
	"log"
	"os"
	"time"
)

type Logger interface {
	Printf(format string, args ...interface{})
}

var (
	logLmsgprefix = 64
	defaultLogger = Logger(log.New(os.Stderr, "[GoroutinePool]: ", log.LstdFlags|logLmsgprefix|log.Lmicroseconds))
)

type Options struct {

	//禁用清理
	DisablePurge bool

	//协程池中的协程如果长时间不执行任务，超过MaxIdleTime就会被清理掉
	MaxIdleTime time.Duration

	// NonBlocking = true 非阻塞模式
	// NonBlocking = false 阻塞模式 下面的MaxBlockingTasks参数才有用
	NonBlocking bool

	//最多阻塞的协程数  为0表示可以无限阻塞
	MaxBlockingTasks int32

	//日志打印器
	Logger Logger

	//协程panic的时候调用
	PanicHandler func(interface{})
}

//使用选项模式，可以方便传参
type Option func(opt *Options)

func SetDisablePurge(disablePurge bool) Option {
	return func(opt *Options) {
		opt.DisablePurge = disablePurge
	}
}

func SetMaxIdleTime(duration time.Duration) Option {
	return func(opt *Options) {
		opt.MaxIdleTime = duration
	}
}

func SetNonBlocking(nonBlocking bool) Option {
	return func(opt *Options) {
		opt.NonBlocking = nonBlocking
	}
}

func SetMaxBlockingTasks(maxBlockingTasks int32) Option {
	return func(opt *Options) {
		opt.MaxBlockingTasks = maxBlockingTasks
	}
}

func SetLogger(logger Logger) Option {
	return func(opt *Options) {
		opt.Logger = logger
	}
}

func SetPanicHandler(panicHandler func(interface{})) Option {
	return func(opt *Options) {
		opt.PanicHandler = panicHandler
	}
}
