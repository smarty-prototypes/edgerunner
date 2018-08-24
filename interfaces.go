package main

import "io"

type Task interface {
	Init() error
	Listen()
	io.Closer
}
type Scheduler interface {
	Schedule()
}

type SignalReader interface {
	Read() bool
}

type TaskFactory func() Task
type SchedulerFactory func(SignalReader) Scheduler
