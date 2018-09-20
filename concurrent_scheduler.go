package edgerunner

import (
	"sync"
	"sync/atomic"
)

type ConcurrentScheduler struct {
	reader   SignalReader
	factory  TaskFactory
	waiter   *sync.WaitGroup
	startup  chan error
	shutdown chan struct{}
	head     *safeTask
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

	proposed, previous := newSafeTask(this.factory(), this.shutdown), this.head
	go this.runTask(proposed, previous)

	if this.err = <-this.startup; this.err != nil {
		proposed.Close()
	} else {
		this.head = proposed
	}

	return this.reader.Read()
}

func (this *ConcurrentScheduler) runTask(proposed, previous *safeTask) {
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

type safeTask struct {
	Task
	shutdown chan<- struct{}
	clean    uint32
}

func newSafeTask(inner Task, shutdown chan<- struct{}) *safeTask {
	return &safeTask{Task: inner, shutdown: shutdown}
}

func (this *safeTask) Listen() {
	this.Task.Listen()
	if atomic.LoadUint32(&this.clean) == 0 {
		this.shutdown <- struct{}{} // didn't shutdown cleanly
	}
}
func (this *safeTask) Close() error {
	if this != nil && atomic.CompareAndSwapUint32(&this.clean, 0, 1) {
		return this.Task.Close()
	}

	return nil
}
