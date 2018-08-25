package edgerunner

type Runner struct {
	signaler Signaler
	factory  SchedulerFactory
}

func NewRunner(signaler Signaler, factory SchedulerFactory) *Runner {
	return &Runner{signaler: signaler, factory: factory,}
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
