package edgerunner

import (
	"io"
	"sync"
)

type ConcurrentScheduler struct {
	reader  SignalReader
	factory TaskFactory
	waiter  *sync.WaitGroup
	mutex   *sync.Mutex
	err     error

	previous chan struct{}
	current  chan struct{}
}

func NewConcurrentScheduler(reader SignalReader, factory TaskFactory) *ConcurrentScheduler {
	return &ConcurrentScheduler{
		reader:  reader,
		factory: factory,
		waiter:  &sync.WaitGroup{},
		mutex:   &sync.Mutex{},
	}
}

func (this *ConcurrentScheduler) Schedule() error {
	for this.schedule() {
	}

	closeChannel(this.current) // signal the current instance to terminate
	this.waiter.Wait()         // wait for all instantiated tasks to 100% completely finish
	return this.loadError()    // did the most recent task have any problem initializing?
}
func (this *ConcurrentScheduler) schedule() bool {
	task := this.prepare()
	go this.runTask(task, this.previous)
	go this.watchSignal(task, this.current)
	return this.reader.Read() // blocking
}
func (this *ConcurrentScheduler) prepare() Task {
	this.waiter.Add(1) // another task has started
	this.previous = this.current
	if this.loadError() == nil {
		// this previous task started properly, so now we can create a new channel
		this.current = make(chan struct{}, 2)
	}

	return this.factory()
}
func (this *ConcurrentScheduler) runTask(task Task, signal chan struct{}) {
	defer this.waiter.Done() // task has finished

	if this.storeError(task.Init()) {
		return // initialize failed
	}

	closeChannel(signal) // signal the previous task (if any) that we're starting to listen
	task.Listen()
}
func (this *ConcurrentScheduler) watchSignal(task io.Closer, signal chan struct{}) {
	<-signal
	task.Close()
}

func (this *ConcurrentScheduler) storeError(err error) bool {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	this.err = err
	return err != nil
}
func (this *ConcurrentScheduler) loadError() error {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	return this.err
}

func closeChannel(channel chan struct{}) {
	if channel != nil {
		close(channel)
	}
}
