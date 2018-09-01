package edgerunner

import (
	"io"
	"sync"
)

type ConcurrentScheduler struct {
	reader   SignalReader
	factory  TaskFactory
	shutdown chan struct{}
	startup  chan error
	waiter   *sync.WaitGroup
	mutex    *sync.Mutex
	active   []io.Closer
	closed   bool
	err      error
}

func NewConcurrentScheduler(reader SignalReader, factory TaskFactory) *ConcurrentScheduler {
	return &ConcurrentScheduler{
		reader:   reader,
		factory:  factory,
		shutdown: make(chan struct{}, 2),
		startup:  make(chan error, 2),
		waiter:   &sync.WaitGroup{},
		mutex:    &sync.Mutex{},
	}
}

func (this *ConcurrentScheduler) Schedule() error {
	go this.schedule()
	<-this.shutdown
	this.cleanup()
	return this.err
}
func (this *ConcurrentScheduler) schedule() {
	for this.scheduleMoreTasks() {
	}

	this.shutdown <- struct{}{}
}

func (this *ConcurrentScheduler) scheduleMoreTasks() bool {
	task := this.factory()

	previous := this.appendActive(task)
	go this.runTask(task, previous)
	if this.err = <-this.startup; this.err != nil {
		this.closeTask(task)
	}

	return this.reader.Read() && !this.isClosed()
}
func (this *ConcurrentScheduler) runTask(task Task, previous io.Closer) {
	defer this.waiter.Done() // when this function exits, the task is considered complete

	if !this.initializeTask(task) {
		return
	}

	if previous != nil {
		go previous.Close() // go vs inline?
	}

	task.Listen()

	// TODO: if this exits early: this.shutdown <- struct{}{}
}
func (this *ConcurrentScheduler) initializeTask(task Task) bool {
	err := task.Init()
	this.startup <- err
	return err == nil
}

func (this *ConcurrentScheduler) appendActive(item io.Closer) (previous io.Closer) {
	this.waiter.Add(1)

	this.mutex.Lock()
	defer this.mutex.Unlock()

	if len(this.active) > 0 {
		previous = this.active[len(this.active)-1]
	}

	this.active = append(this.active, item)
	return previous
}
func (this *ConcurrentScheduler) closeTask(task io.Closer) {
	task.Close()

	this.mutex.Lock()
	defer this.mutex.Unlock()

	if len(this.active) > 0 {
		this.active = this.active[:len(this.active)-1] // remove the last item
	}
}

func (this *ConcurrentScheduler) cleanup() {
	this.shutdownActive()
	this.waiter.Wait()
	close(this.startup)
	close(this.shutdown)
}
func (this *ConcurrentScheduler) shutdownActive() {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	if this.closed {
		return
	}

	for _, item := range this.active {
		item.Close()
	}

	this.closed = true
	this.active = nil
}
func (this *ConcurrentScheduler) isClosed() bool {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	return this.closed
}
