package edgerunner

type Runner interface {
	Start()
	Stop()
	Reload()
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
	SchedulerFactory func(Reader) Scheduler
)

type (
	Signaler interface {
		Start() (Reader, bool)
		Stop()
		Signal() bool
	}
	Reader interface {
		Read() bool
	}
)
