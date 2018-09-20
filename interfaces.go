package edgerunner

type Runner interface {
	Start() error
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
		Schedule() error
	}
	SchedulerFactory func(SignalReader) Scheduler
)

type (
	Signaler interface {
		Start() SignalReader
		Stop()
		Signal() bool
	}
	SignalReader interface {
		Read() bool
	}
)
