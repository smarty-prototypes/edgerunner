package edgerunner

import "sync"

type ConcurrentScheduler struct {
	reader        SignalReader
	factory       TaskFactory
	waiter        *sync.WaitGroup
	startupResult chan error
	head          Task
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
	this.waiter.Add(1)

	proposed, previous := this.factory(), this.head
	go this.runTask(proposed, previous)

	if this.err = <-this.startupResult; this.err != nil {
		proposed.Close()
	} else {
		this.head = proposed
	}

	return this.reader.Read()
}

func (this *ConcurrentScheduler) runTask(proposed, previous Task) {
	defer this.waiter.Done()
	if !this.initializeTask(proposed) {
		return
	}

	if previous != nil {
		previous.Close() // note: semi-concurrent = previous.Close(), fully concurrent = go previous.Close()
	}

	proposed.Listen()
}
func (this *ConcurrentScheduler) initializeTask(task Task) bool {
	err := task.Init()
	this.startupResult <- err
	return err == nil
}
