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
	var previous, current chan struct{}

	for {
		this.waiter.Add(1) // another task has started

		previous = current
		if this.loadError() == nil {
			// this previous task started properly, so now we can create a new channel
			current = make(chan struct{}, 2)
		}

		task := this.factory()
		go this.runTask(task, previous)
		go this.watchSignal(task, current)

		if !this.reader.Read() {
			break
		}
	}

	close(current)
	this.waiter.Wait() // wait for all instantiated tasks to 100% completely finish
	return this.loadError()
}
func (this *ConcurrentScheduler) runTask(task Task, signal chan struct{}) {
	defer this.waiter.Done() // task has finished

	if this.storeError(task.Init()) {
		return // initialize failed
	}

	close(signal) // signal the previous task (if any) that we're starting to listen
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
