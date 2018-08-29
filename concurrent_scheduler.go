package edgerunner

import (
	"sync"
	"sync/atomic"
)

type ConcurrentScheduler struct {
	reader         SignalReader
	factory        TaskFactory
	again          uint32
	osSignalWaiter *sync.WaitGroup
	previousTask   Task
}

func NewConcurrentScheduler(reader SignalReader, factory TaskFactory) *ConcurrentScheduler {
	return &ConcurrentScheduler{
		reader:         reader,
		factory:        factory,
		osSignalWaiter: &sync.WaitGroup{},
	}
}

func (this *ConcurrentScheduler) Schedule() error {
	atomic.StoreUint32(&this.again, 1)

	for this.canScheduleAgain() {
		this.osSignalWaiter.Add(1)
		go this.scheduleTask()
		this.osSignalWaiter.Wait()
		if this.previousTask != nil {
			this.previousTask.Close()
		}
	}

	return nil
}

func (this *ConcurrentScheduler) scheduleTask() bool {
	task := this.factory()
	go this.watchSignal(task)
	task.Init() // TODO: if error

	// TODO: how to delay this until the current task is fully listening
	if this.previousTask != nil {
		this.previousTask.Close()
	}
	task.Listen() // blocks until closed, but cannot close until the next task is listening

	return this.canScheduleAgain()
}

func (this *ConcurrentScheduler) watchSignal(task Task) {
	this.scheduleAgain(this.reader.Read())
	this.previousTask = task
	this.osSignalWaiter.Done()
}

func (this *ConcurrentScheduler) canScheduleAgain() bool {
	defer atomic.StoreUint32(&this.again, 0) // reset state
	return atomic.LoadUint32(&this.again) == 1
}

func (this *ConcurrentScheduler) scheduleAgain(again bool) {
	if again {
		atomic.StoreUint32(&this.again, 1)
	} else {
		atomic.StoreUint32(&this.again, 0)
	}
}
