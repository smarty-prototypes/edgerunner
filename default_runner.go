package edgerunner

type DefaultRunner struct {
	signaler Signaler
	factory  SchedulerFactory
}

func NewRunner(signaler Signaler, factory SchedulerFactory) Runner {
	return &DefaultRunner{signaler: signaler, factory: factory}
}

func (this *DefaultRunner) Start() {
	if reader, started := this.signaler.Start(); started {
		scheduler := this.factory(reader)
		scheduler.Schedule()
		this.signaler.Stop()
	}
}

func (this *DefaultRunner) Stop() {
	this.signaler.Stop()
}

func (this *DefaultRunner) Reload() bool {
	return this.signaler.Signal()
}
