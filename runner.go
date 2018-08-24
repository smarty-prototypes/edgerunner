package main

import (
	"sync"
)

// TODO: put all complexity of mutex lock/unlock, etc.
// behind another instance
// make signal watcher more complex and have it return a struct
// that has a Continue() bool func on it...

type Runner struct {
	factory SchedulerFactory
	signals chan interface{}
	mutex   *sync.Mutex
}

func NewRunner(factory SchedulerFactory) *Runner {
	return &Runner{factory: factory, mutex: &sync.Mutex{}}
}

func (this *Runner) Start() {
	watcher := this.start()
	scheduler := this.factory(watcher)
	scheduler.Schedule()

	this.Stop() // in case schedule exits without stop being called
}
func (this *Runner) start() SignalReader {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	if this.signals == nil {
		this.signals = make(chan interface{}, 2)
	}

	return NewChannelWatcher(this.signals)
}

func (this *Runner) Stop() {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	if this.signals != nil {
		close(this.signals)
		this.signals = nil
	}
}

func (this *Runner) Reload() {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	if this.signals != nil && len(this.signals) == 0 {
		this.signals <- nil
	}
}
