package main

type Runner struct {
	factory  SchedulerFactory
	signaler Signaler
	running  bool
}

func NewRunner(factory SchedulerFactory, signaler Signaler) *Runner {
	return &Runner{factory: factory, signaler: signaler}
}

func (this *Runner) Start() {
	if this.running {
		return
	}
	this.running = true
	reader := this.signaler.Start()
	scheduler := this.factory(reader)
	scheduler.Schedule()

	this.Stop() // in case schedule exits without stop being called
	this.running = false
}

func (this *Runner) Stop() {
	this.signaler.Stop()
}

func (this *Runner) Reload() bool {
	return this.signaler.Signal()
}
