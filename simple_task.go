package main

type SerialTask struct {
	factory func() Task
	task    Task
}

func NewSerialTask(factory func() Task) *SerialTask {
	return &SerialTask{factory: factory}
}

func (this *SerialTask) RunLoop() {
	this.task = this.factory()
	this.task.Init()
	this.task.Listen()
}

func (this *SerialTask) ShutdownLoop() {
	this.task.Close()
}

//////////////////////////////////////////

type ConcurrentTask struct {
	channel chan Task
	factory func() Task
}

func NewConcurrenTask(factory func() Task) *ConcurrentTask {
	return &ConcurrentTask{
		factory: factory,
		channel: make(chan Task, 2),
	}
}

func (this *ConcurrentTask) RunLoop() {
	task := this.factory()
	this.channel <- task
	task.Init()
	task.Listen()
}

func (this *ConcurrentTask) ShutdownLoop() {
	(<-this.channel).Close()
}

