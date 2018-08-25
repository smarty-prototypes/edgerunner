package edgerunner

import (
	"sync/atomic"
)

type SerialScheduler struct {
	reader  SignalReader
	factory TaskFactory
	again   uint32
}

func NewSerialScheduler(reader SignalReader, factory TaskFactory) *SerialScheduler {
	return &SerialScheduler{
		reader:  reader,
		factory: factory,
	}
}

func (this *SerialScheduler) Schedule() {
	for this.scheduleTask() {
	}
}

func (this *SerialScheduler) scheduleTask() bool {
	task := this.factory()
	go this.watchSignal(task)
	task.Init()   // TODO: if error
	task.Listen() // we only schedule again if listen exits correctly

	return this.canScheduleAgain()
}

func (this *SerialScheduler) watchSignal(task Task) {
	this.scheduleAgain(this.reader.Read())
	task.Close()
}

func (this *SerialScheduler) canScheduleAgain() bool {
	defer atomic.StoreUint32(&this.again, 0) // reset state
	return atomic.LoadUint32(&this.again) == 1
}

func (this *SerialScheduler) scheduleAgain(again bool) {
	if again {
		atomic.StoreUint32(&this.again, 1)
	} else {
		atomic.StoreUint32(&this.again, 0)
	}
}
