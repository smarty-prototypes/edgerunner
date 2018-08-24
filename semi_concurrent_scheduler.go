package main

import "sync"

type SemiConcurrentScheduler struct {
	signals <-chan interface{}
	factory TaskFactory
	waiter  *sync.WaitGroup
}

func NewSemiConcurrentScheduler(signals <-chan interface{}, factory TaskFactory) *SemiConcurrentScheduler {
	return &SemiConcurrentScheduler{
		signals: signals,
		factory: factory,
		waiter:  &sync.WaitGroup{},
	}
}

func (this *SemiConcurrentScheduler) Schedule() {
	// TODO
}
