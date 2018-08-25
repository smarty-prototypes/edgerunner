package main

type Runner struct {
	factory  SchedulerFactory
	signaler Signaler
}

func NewRunner(factory SchedulerFactory, signaler Signaler) *Runner {
	return &Runner{factory: factory, signaler: signaler}
}

func (this *Runner) Start() {
	reader, started := this.signaler.Start()
	if !started {
		return
	}

	scheduler := this.factory(reader)
	scheduler.Schedule()

	this.Stop() // in case schedule exits without stop being called
}

func (this *Runner) Stop() {
	this.signaler.Stop()
}

func (this *Runner) Reload() bool {
	return this.signaler.Signal()
}
