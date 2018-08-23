package main

import (
	"sync"
)

type Runner struct {
	mutex   *sync.Mutex
	signals chan interface{}
	factory func() Task
	active  []Task
}

func NewRunner(factory func() Task) *Runner {
	return &Runner{
		mutex:   &sync.Mutex{},
		factory: factory,
	}
}

func (this *Runner) Start() {
	if !this.tryStart() {
		return
	}

	for this.isStarted() {
		instance := this.newTask()
		instance.Init()
		instance.Listen()
	}
}
func (this *Runner) Stop() {
	if !this.isStarted() {
		return
	}

	this.mutex.Lock()
	defer this.mutex.Unlock()
	this.sendSignal()
	close(this.signals)
	this.signals = nil
}
func (this *Runner) Reload() {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	this.sendSignal()
}

func (this *Runner) newTask() Task {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	instance := this.factory()
	this.active = append(this.active, instance)
	return instance
}
func (this *Runner) sendSignal() {
	if this.signals == nil {
		return
	}

	select {
	case this.signals <- nil: // try and send a signal if the channel can hold it
	default:
	}
}
func (this *Runner) tryStart() bool {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	if this.signals != nil {
		return false // already created and running, don't do it again
	}

	// signal needs to be created and watched
	this.signals = make(chan interface{}, 4)
	go this.watch(this.signals)
	return true
}
func (this *Runner) watch(signals <-chan interface{}) {
	for range signals {
		if len(signals) == 0 {
			this.closeActive()
		}
	}
}
func (this *Runner) closeActive() {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	for _, item := range this.active {
		item.Close()
	}

	this.active = nil
}
func (this *Runner) isStarted() bool {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	return this.signals != nil
}
