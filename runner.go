package edgerunner

type Runner struct {
	factory  SchedulerFactory
	signaler Signaler
}

func NewRunner(factory SchedulerFactory, signaler Signaler) *Runner {
	return &Runner{factory: factory, signaler: signaler}
}

func (this *Runner) Start() {
	if reader, started := this.signaler.Start(); started {
		scheduler := this.factory(reader)
		scheduler.Schedule()
		this.signaler.Stop()
	}
}

func (this *Runner) Stop() {
	this.signaler.Stop()
}

func (this *Runner) Reload() bool {
	return this.signaler.Signal()
}
