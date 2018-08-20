package main

import (
	"fmt"
	"sync"
)

type SampleApp struct {
	waiter *sync.WaitGroup
}

func NewApp() Runnable {
	waiter := &sync.WaitGroup{}
	waiter.Add(1)

	return &SampleApp{
		waiter: waiter,
	}
}
func (this *SampleApp) Initialize() error {
	fmt.Println("Initializing...")
	return nil
}
func (this *SampleApp) Listen() {
	fmt.Println("Listening...")
	this.waiter.Wait()
}
func (this *SampleApp) Close() error {
	fmt.Println("Closing...")
	this.waiter.Done()
	return nil
}
func (this *SampleApp) Concurrency() int {
	return -1
}
