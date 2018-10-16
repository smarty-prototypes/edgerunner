package edgerunner

import (
	"sync"
)

type ConcurrentScheduler struct {
	reader   SignalReader
	factory  TaskFactory
	waiter   *sync.WaitGroup
	startup  chan error
	shutdown chan struct{}
	head     *SignalingTask
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

	proposed, previous := NewSignalingTask(this.factory(), this.startup, this.shutdown), this.head
	go this.runTask(proposed, previous)

	if this.err = <-this.startup; this.err != nil {
		proposed.Close()
	} else {
		this.head = proposed
	}

	return this.reader.Read() // wait for some kind of signal
}

func (this *ConcurrentScheduler) runTask(proposed, previous *SignalingTask) {
	defer this.waiter.Done()
	if proposed.Init() != nil {
		return
	}

	if previous != nil {
		previous.Close()
	}

	proposed.Listen()
}
