package main

import (
	"fmt"
	"sync"
)

type SampleApp struct {
	waiter *sync.WaitGroup
}

func NewApp() Task {
	waiter := &sync.WaitGroup{}
	waiter.Add(1)

	return &SampleApp{waiter: waiter}
}

func (this *SampleApp) Init() error {
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
