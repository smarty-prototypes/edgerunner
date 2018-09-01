package edgerunner

import (
	"io"
	"sync"
	"sync/atomic"
)

type ConcurrentScheduler struct {
	reader        SignalReader
	factory       TaskFactory
	waiter        *sync.WaitGroup
	startupResult chan error
	head          *SafeTask
	counter       uint32
	err           error
}

func NewConcurrentScheduler(reader SignalReader, factory TaskFactory) *ConcurrentScheduler {
	return &ConcurrentScheduler{
		reader:        reader,
		factory:       factory,
		waiter:        &sync.WaitGroup{},
		startupResult: make(chan error, 2),
	}
}

func (this *ConcurrentScheduler) Schedule() error {
	for this.schedule() {
	}

	this.head.Close()
	this.waiter.Wait()

	return this.err
}
func (this *ConcurrentScheduler) schedule() bool {
	if !this.head.IsSafe() {
		return false
	}

	this.waiter.Add(1)
	task := &SafeTask{Task: this.factory()}
	previous := this.head
	go this.runTask(task, previous)

	if this.err = <-this.startupResult; this.err != nil {
		task.Close()
	} else {
		this.head = task
	}

	if this.counter == 0 && this.err != nil {
		return false
	}

	this.counter++
	return this.reader.Read() && this.head.IsSafe()
}

func (this *ConcurrentScheduler) runTask(proposed Task, previous io.Closer) {
	defer this.waiter.Done()
	if !this.initializeTask(proposed) {
		return
	}

	// note: semi-concurrent = previous.Close(), fully concurrent = go previous.Close()
	previous.Close()
	proposed.Listen()

}
func (this *ConcurrentScheduler) initializeTask(task Task) bool {
	err := task.Init()
	this.startupResult <- err
	return err == nil
}

/////////////////////////////////////////////////////////////////

type SafeTask struct {
	Task
	state uint32
}

func (this *SafeTask) Close() error {
	if this == nil {
		return nil
	} else if atomic.CompareAndSwapUint32(&this.state, 0, 1) {
		return this.Task.Close()
	} else {
		panic("task already closed")
	}
}
func (this *SafeTask) Listen() {
	this.Task.Listen()
	atomic.CompareAndSwapUint32(&this.state, 0, 2)
}
func (this *SafeTask) IsSafe() bool {
	return this == nil || atomic.LoadUint32(&this.state) != 2
}
