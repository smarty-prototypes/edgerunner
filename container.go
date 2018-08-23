package main

import (
	"sync"
)

type Container struct {
	factory      func() Task
	channel      chan interface{}
	startWaiter  *sync.WaitGroup
	listenWaiter *sync.WaitGroup
}

func (this *Container) Start(factory func() Task) {
	this.factory = factory
	this.startWaiter.Add(1)
	go this.start()
	this.startWaiter.Wait()
}
func (this *Container) start() {
	task := this.newTask(nil)

	for range this.channel {
		if len(this.channel) == 0 {
			task = this.newTask(task)
		}
	}

	task.Close()
	this.listenWaiter.Wait()

	this.startWaiter.Done()
}

func (this *Container) newTask(task Task) Task {
	if task != nil {
		task.Close()
		this.listenWaiter.Wait()
	}

	task = this.factory()
	this.listenWaiter.Add(1)

	go func() {
		task.Init()
		task.Listen()
		this.listenWaiter.Done()
	}()

	return task
}

func (this *Container) Reload() {
	this.channel <- nil
}
func (this *Container) Close() {
	// TODO: mutex around this guy
	close(this.channel)
}
