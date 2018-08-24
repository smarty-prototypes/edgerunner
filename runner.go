package main

import "sync"

type Runner struct {
	mutex   *sync.Mutex
	waiter  *sync.WaitGroup
	channel chan interface{}
	factory SchedulerFactory
}

func NewRunner(factory SchedulerFactory) *Runner {
	return &Runner{
		mutex:   &sync.Mutex{},
		waiter:  &sync.WaitGroup{},
		factory: factory,
	}
}

func (this *Runner) Start() {
	this.start()
	this.waiter.Wait()
}
func (this *Runner) start() {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	if this.channel != nil {
		return
	}

	this.waiter.Add(1)
	this.channel = make(chan interface{}, 2)
	channel := this.channel
	go this.schedule(channel) // use local copy for closure operation
}
func (this *Runner) schedule(channel <-chan interface{}) {
	scheduler := this.factory(channel)
	scheduler.Schedule()
	this.waiter.Done()
}

func (this *Runner) Stop() {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	if this.channel == nil {
		return
	}

	close(this.channel)
	this.channel = nil
}

func (this *Runner) Reload() {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	if this.channel == nil {
		return // we're stopped
	}

	if len(this.channel) > 0 {
		return // the channel already has a reload signal on it
	}

	select {
	case this.channel <- nil: // send the reload signal
	default:
		// the channel is full, there's already a reload signal on it
	}
}
