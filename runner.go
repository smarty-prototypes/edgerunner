package main

import (
	"fmt"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
)

type Runner struct {
	state    uint32
	factory  func() Runnable
	instance Runnable
}

func NewRunner(factory func() Runnable) *Runner {
	return &Runner{ factory: factory }
}

func (this *Runner) Start() error {
	if !atomic.CompareAndSwapUint32(&this.state, 0, 1) {
		return
	}

	go func() {
		terminate := make(chan os.Signal, 16)
		signal.Notify(terminate, os.Interrupt)

		fmt.Printf("\nReceived shutdown signal [%s]\n", <-terminate)
		close(terminate)
		this.Stop()

		signal.Stop(terminate)
	}()

	go func() {
		sighup := make(chan os.Signal, 16)
		signal.Notify(sighup, syscall.SIGHUP)

		for item := range sighup {
			fmt.Printf("\nReceived reload signal [%s]\n", item)
			this.closeActive()
		}
	}()

	for this.isStarted() {
		this.instance = this.factory()
		err := this.instance.Initialize()
		if err != nil {
			return err
		}
		this.instance.Listen()
	}
	return nil
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
