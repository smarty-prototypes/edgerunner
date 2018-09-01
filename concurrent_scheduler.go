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
	head     *ConcurrentTask
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
	go func() {
		for this.schedule() {
		}

		this.shutdown <- struct{}{} // signal the shutdown channel that we're ready to close
	}()

	<-this.shutdown
	this.cleanup()
	return this.err
}
func (this *ConcurrentScheduler) schedule() bool {
	this.waiter.Add(1)

	proposed := &ConcurrentTask{Task: this.factory(), signal: this.shutdown}
	go this.runTask(proposed, this.head)

	if this.err = <-this.startup; this.err != nil {
		proposed.Close()
	} else {
		this.head = proposed
	}

	return this.reader.Read()
}

func (this *ConcurrentScheduler) runTask(proposed, previous Task) {
	defer this.waiter.Done()
	if this.initializeTask(proposed) {
		previous.Close() // note: semi-concurrent = previous.Close(), fully concurrent = go previous.Close()
		proposed.Listen()
	}
}
func (this *ConcurrentScheduler) initializeTask(task Task) bool {
	err := task.Init()
	this.startup <- err
	return err == nil
}

func (this *ConcurrentScheduler) cleanup() {
	this.head.Close()
	this.waiter.Wait()
	close(this.shutdown)
	close(this.startup)
}

///////////////////////////////////////////////////////////////////////

type ConcurrentTask struct {
	Task
	signal chan struct{}
	state  uint32
}

func (this *ConcurrentTask) Listen() {
	this.Task.Listen()
	if !atomic.CompareAndSwapUint32(&this.state, 1, 2) {
		this.signal <- struct{}{} // didn't shutdown cleanly
	}
}
func (this *ConcurrentTask) Close() error {
	if this != nil && atomic.CompareAndSwapUint32(&this.state, 0, 1) {
		return this.Task.Close()
	}

	return nil
}
