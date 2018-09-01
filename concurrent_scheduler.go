package edgerunner

import (
	"io"
	"sync"
	"sync/atomic"
)

type ConcurrentScheduler struct {
	reader   SignalReader
	factory  TaskFactory
	waiter   *sync.WaitGroup
	startup  chan error
	shutdown chan struct{}
	head     io.Closer
	err      error
}

func NewConcurrentScheduler(reader SignalReader, factory TaskFactory) *ConcurrentScheduler {
	return &ConcurrentScheduler{
		reader:   reader,
		factory:  factory,
		waiter:   &sync.WaitGroup{},
		startup:  make(chan error, 2),
		shutdown: make(chan struct{}, 2),
	}
}

func (this *ConcurrentScheduler) Schedule() error {
	go this.scheduleTasks()
	<-this.shutdown
	this.head.Close()
	this.waiter.Wait()
	close(this.shutdown)
	close(this.startup)
	return this.err
}
func (this *ConcurrentScheduler) scheduleTasks() {
	for this.scheduleNextTask() {
	}

	this.shutdown <- struct{}{} // signal the shutdown channel that we're ready to close
}
func (this *ConcurrentScheduler) scheduleNextTask() bool {
	this.waiter.Add(1)

	proposed, previous := &ConcurrentTask{Task: this.factory(), shutdown: this.shutdown}, this.head
	go this.runTask(proposed, previous)

	if this.err = <-this.startup; this.err != nil {
		proposed.Close()
	} else {
		this.head = proposed
	}

	return this.reader.Read()
}

func (this *ConcurrentScheduler) runTask(proposed Task, previous io.Closer) {
	defer this.waiter.Done()
	if this.initializeTask(proposed) {
		previous.Close()
		proposed.Listen()
	}
}
func (this *ConcurrentScheduler) initializeTask(task Task) bool {
	err := task.Init()
	this.startup <- err
	return err == nil
}

///////////////////////////////////////////////////////////////////////

type ConcurrentTask struct {
	Task
	shutdown chan struct{}
	clean    uint32
}

func (this *ConcurrentTask) Listen() {
	this.Task.Listen()
	if atomic.LoadUint32(&this.clean) == 0 {
		this.shutdown <- struct{}{} // didn't shutdown cleanly
	}
}
func (this *ConcurrentTask) Close() error {
	if this != nil && atomic.CompareAndSwapUint32(&this.clean, 0, 1) {
		return this.Task.Close()
	}

	return nil
}
