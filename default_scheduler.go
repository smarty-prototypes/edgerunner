package edgerunner

import "sync"

type DefaultScheduler struct {
	reader   SignalReader
	factory  TaskFactory
	waiter   *sync.WaitGroup
	startup  chan error
	shutdown chan struct{}
	head     *safeTask
	err      error
}

func NewDefaultScheduler(reader SignalReader, factory TaskFactory) *DefaultScheduler {
	return &DefaultScheduler{
		reader:   reader,
		factory:  factory,
		waiter:   &sync.WaitGroup{},
		startup:  make(chan error, 2),
		shutdown: make(chan struct{}, 2),
	}
}

func (this *DefaultScheduler) Schedule() error {
	go this.scheduleTasks()
	<-this.shutdown
	this.cleanup()
	return this.err
}

func (this *DefaultScheduler) scheduleTasks() {
	for this.scheduleNextTask() {
	}

	this.shutdown <- struct{}{} // signal the shutdown channel that we're ready to close
}

func (this *DefaultScheduler) scheduleNextTask() bool {
	this.waiter.Add(1)

	proposed := newSafeTask(this.factory(), this.shutdown)
	go this.runTask(proposed, this.head)

	if this.err = <-this.startup; this.err != nil {
		proposed.Close()
	} else {
		this.head = proposed
	}

	return this.reader.Read()
}
func (this *DefaultScheduler) runTask(proposed, previous *safeTask) {
	// TODO
}

func (this *DefaultScheduler) cleanup() {
	if this.head != nil {
		this.head.Close()
		this.head = nil
	}
	this.waiter.Wait()
	close(this.shutdown)
	close(this.startup)
}
