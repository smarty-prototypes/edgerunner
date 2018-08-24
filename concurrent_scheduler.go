package main

import "sync"

type ConcurrentScheduler struct {
	signals <-chan interface{}
	factory TaskFactory
	waiter  *sync.WaitGroup
}

func NewConcurrentScheduler(signals <-chan interface{}, factory TaskFactory) *ConcurrentScheduler {
	return &ConcurrentScheduler{
		signals: signals,
		factory: factory,
		waiter:  &sync.WaitGroup{},
	}
}

func (this *ConcurrentScheduler) Schedule() {
	// TODO
}
