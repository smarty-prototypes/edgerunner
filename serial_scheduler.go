package edgerunner

import "sync/atomic"

type SerialScheduler struct {
	reader  SignalReader
	factory TaskFactory
	again   uint32
	err     error
}

func NewSerialScheduler(reader SignalReader, factory TaskFactory) *SerialScheduler {
	return &SerialScheduler{
		reader:  reader,
		factory: factory,
	}
}

func (this *SerialScheduler) Schedule() error {
	for this.scheduleTask() {
	}
	return this.err
}

func (this *SerialScheduler) scheduleTask() bool {
	task := this.factory()
	go this.watchSignal(task)

	if this.err = task.Init(); this.err != nil {
		return false
	} else {
		task.Listen() // we only schedule again if listen exits correctly
	}

	return this.canScheduleAgain()
}

func (this *SerialScheduler) watchSignal(task Task) {
	reload := this.reader.Read()
	err := task.Close()
	this.scheduleAgain(reload && err == nil)
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
