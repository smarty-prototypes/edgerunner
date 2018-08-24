package main

import "sync"

type ConcurrentScheduler struct {
	signals <-chan interface{}
	factory TaskFactory
	waiter  *sync.WaitGroup
}

func NewConcurrentScheduler(signals <-chan interface{}, factory TaskFactory) *ConcurrentScheduler {
	return &ConcurrentScheduler{
		signals: signals,
		factory: factory,
		waiter:  &sync.WaitGroup{},
	}
}

func (this *ConcurrentScheduler) Schedule() {
	for this.scheduleTask() {
	}
}

func (this *ConcurrentScheduler) scheduleTask() bool {
	this.waiter.Add(1)
	task := this.factory()

	go this.runTask(task)
	defer this.closeTask(task)

	return this.readSignal()
}

func (this *ConcurrentScheduler) runTask(task Task) {
	task.Init()
	task.Listen()

	this.waiter.Done()
}

func (this *ConcurrentScheduler) closeTask(task Task) {
	task.Close()

	this.waiter.Wait()
}

func (this *ConcurrentScheduler) readSignal() (ok bool) {
	//	for len(this.signals) > 0 {
	_, ok = <-this.signals
	//	}
	return ok
}
