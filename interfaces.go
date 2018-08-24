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

type TaskFactory func() Task
type SchedulerFactory func(<-chan interface{}) Scheduler
