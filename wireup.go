package edgerunner

import "os"

func Configure(factory TaskFactory, options ...Option) Runner {
	this := &wireupState{
		signaler:    NewSignaler(),
		reload:      DefaultReloadSignals,
		terminate:   DefaultShutdownSignals,
		concurrency: 0, // sequential
	}

	for _, option := range options {
		option(this) // override defaults here
	}

	if len(this.reload) == 0 || len(this.terminate) == 0 {
		panic("no signals to watch")
	}

	runner := NewRunner(this.signaler, func(reader SignalReader) Scheduler {
		return this.Scheduler(reader, factory)
	})
	return NewSignalRunner(runner, DefaultReloadSignals, DefaultShutdownSignals)
}

type Option func(this *wireupState)

func WatchReloadSignals(signals ...os.Signal) Option {
	return func(this *wireupState) { this.reload = signals }
}
func WatchShutdownSignals(signals ...os.Signal) Option {
	return func(this *wireupState) { this.terminate = signals }
}
func Sequential() Option {
	return func(this *wireupState) { this.concurrency = 0 }
}
func Semiconcurrent() Option {
	panic("not yet supported")
	return func(this *wireupState) { this.concurrency = 1 }
}
func Concurrent() Option {
	return func(this *wireupState) { this.concurrency = 2 }
}

type wireupState struct {
	signaler    Signaler
	reload      []os.Signal
	terminate   []os.Signal
	concurrency int
}

func (this *wireupState) Scheduler(reader SignalReader, factory TaskFactory) Scheduler {
	switch this.concurrency {
	case 2:
		return NewConcurrentScheduler(reader, factory)
	case 1:
		panic("not yet supported")
	case 0:
		return NewSerialScheduler(reader, factory)
	default:
		panic("concurrency level not supported")
	}
}
