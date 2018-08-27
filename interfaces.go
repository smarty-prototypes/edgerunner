package edgerunner

type Runner interface {
	Start()
	Stop()
	Reload() bool
}

type (
	Task interface {
		Init() error
		Listen()
		Close() error
	}
	TaskFactory func() Task
)

type (
	Scheduler interface {
		Schedule()
	}
	SchedulerFactory func(SignalReader) Scheduler
)

type (
	Signaler interface {
		Start() (SignalReader, bool)
		Stop()
		Signal() bool
	}
	SignalReader interface {
		Read() bool
	}
)
