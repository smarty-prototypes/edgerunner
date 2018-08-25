package edgerunner

import (
	"sync"
	"sync/atomic"
)

type ConcurrentScheduler struct {
	reader       SignalReader
	factory      TaskFactory
	again        uint32
	signal       *sync.WaitGroup
	previousTask Task
}

func NewConcurrentScheduler(reader SignalReader, factory TaskFactory) *ConcurrentScheduler {
	return &ConcurrentScheduler{
		reader:  reader,
		factory: factory,
		signal:  &sync.WaitGroup{},
	}
}

func (this *ConcurrentScheduler) Schedule() {

	for {
		this.signal.Add(1)
		go this.scheduleTask()
		this.signal.Wait()
		if !this.canScheduleAgain() {
			break
		}
	}

}

func (this *ConcurrentScheduler) scheduleTask() bool {
	task := this.factory()
	go this.watchSignal(task)
	task.Init() // TODO: if error

	// TODO: shouldn't do this until this task is fully listening
	if this.previousTask != nil {
		this.previousTask.Close()
	}
	task.Listen() // blocks until closed, but cannot close until the next task is listening

	return this.canScheduleAgain()
}

func (this *ConcurrentScheduler) watchSignal(task Task) {
	this.previousTask = task
	this.scheduleAgain(this.reader.Read()) // blocks until the channel is populated or closed
	this.signal.Done()
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
