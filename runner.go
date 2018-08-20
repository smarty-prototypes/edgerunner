package main

import (
	"fmt"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
)

type Runner struct {
	signals  chan os.Signal
	state    uint32
	factory  func() Runnable
	instance Runnable
}

func NewRunner(factory func() Runnable) *Runner {
	return &Runner{
		factory: factory,
		signals: make(chan os.Signal, 16),
	}
}

func (this *Runner) Start() {
	if !atomic.CompareAndSwapUint32(&this.state, 0, 1) {
		return
	}

	go func() {
		signal.Notify(this.signals, os.Interrupt)
		fmt.Printf("\nReceived shutdown signal [%s]\n", <-this.signals)
		this.Stop()
	}()

	go func() {
		sighup := make(chan os.Signal, 2)
		signal.Notify(sighup, syscall.SIGHUP)

		for item := range sighup {
			fmt.Printf("\nReceived reload signal [%s]\n", item)
			this.closeActive()
		}
	}()

	for this.isStarted() {
		this.instance = this.factory()
		this.instance.Initialize()
		this.instance.Listen()
	}

	signal.Stop(this.signals)
	close(this.signals)
}

func (this *Runner) isStarted() bool {
	return atomic.LoadUint32(&this.state) == 1
}

func (this *Runner) Stop() {
	if atomic.CompareAndSwapUint32(&this.state, 1, 0) {
		this.closeActive()
	}
}

func (this *Runner) closeActive() {
	this.instance.Close()
}
