package edgerunner

import (
	"io"
	"sync"
)

// Issues:
// 1. if Listen() exits without Close() being called
// 2. if Init() returns an error

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
	this.waiter.Add(1) // another task has started

	//	this.previous = this.current
	if this.loadError() == nil {
		this.current = make(chan struct{}, 2) // the previous task started successfully
	}
	task := this.factory()

	go this.runTask(task, this.previous, this.current)
	go this.watchSignal(task, this.current)

	return this.reader.Read() // blocking
}
func (this *ConcurrentScheduler) runTask(task Task, previous, current chan struct{}) {
	defer this.waiter.Done() // task has finished

	if this.storeError(task.Init()) {
		return // initialization failed; don't close the previous task
	}

	// TODO: safe locking
	this.previous = current

	closeChannel(previous) // signal the previous task (if any) that we're starting to listen
	task.Listen()
	// TODO: what happens when task.Listen() exits without task.Close()?
	// if so, we need a way to signal main loop about what's happening...
	// this probably means we pull the main loop out one more layer
	// and have a signal for it that we can make here
	// because if a given task.Listen() fails without close being called, we need
	// to shut down
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
