package main

import "sync"

type SerialScheduler struct {
	signals <-chan interface{}
	factory TaskFactory
	waiter  *sync.WaitGroup
}

func NewSerialScheduler(signals <-chan interface{}, factory TaskFactory) *SerialScheduler {
	return &SerialScheduler{
		signals: signals,
		factory: factory,
		waiter:  &sync.WaitGroup{},
	}
}

func (this *SerialScheduler) Schedule() {
	for this.scheduleTask() {
	}
}

func (this *SerialScheduler) scheduleTask() bool {
	this.waiter.Add(1)
	task := this.factory()

	go this.runTask(task)
	defer this.closeTask(task)

	return this.readSignal()
}

func (this *SerialScheduler) runTask(task Task) {
	task.Init()
	task.Listen()

	this.waiter.Done()
}

func (this *SerialScheduler) closeTask(task Task) {
	task.Close()

	this.waiter.Wait()
}

func (this *SerialScheduler) readSignal() (ok bool) {
//	for len(this.signals) > 0 {
		_, ok = <-this.signals
//	}
	return ok
}
