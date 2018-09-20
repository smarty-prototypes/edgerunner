package edgerunner

type DefaultRunner struct {
	signaler Signaler
	factory  SchedulerFactory
}

func NewRunner(signaler Signaler, factory SchedulerFactory) Runner {
	return &DefaultRunner{signaler: signaler, factory: factory}
}

func (this *DefaultRunner) Start() error {
	if reader := this.signaler.Start(); reader != nil {
		defer this.signaler.Stop()
		scheduler := this.factory(reader)
		return scheduler.Schedule()
	}

	return nil
}

func (this *DefaultRunner) Stop() {
	this.signaler.Stop()
}

func (this *DefaultRunner) Reload() bool {
	return this.signaler.Signal()
}
