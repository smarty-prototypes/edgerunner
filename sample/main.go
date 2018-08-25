package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"

	"github.com/smartystreets/edgerunner"
)

func main() {
	runner := edgerunner.NewRunner(edgerunner.NewSignaler(), newScheduler)

	go func() {
		signals := make(chan os.Signal, 16)
		signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
		fmt.Println("\nSignal received:", <-signals)
		runner.Stop()
	}()

	go func() {
		signals := make(chan os.Signal, 16)
		signal.Notify(signals, syscall.SIGHUP)
		for message := range signals {
			fmt.Println("\nSignal received:", message)
			runner.Reload()
		}
	}()

	runner.Start()
}

func newScheduler(reader edgerunner.Reader) edgerunner.Scheduler {
	return edgerunner.NewSerialScheduler(reader, NewSimpleTask)
}

type SimpleTask struct {
	id     uint32
	waiter *sync.WaitGroup
}

var counter uint32

func NewSimpleTask() edgerunner.Task {
	id := atomic.AddUint32(&counter, 1)
	return &SimpleTask{id: id, waiter: &sync.WaitGroup{}}
}

func (this *SimpleTask) Init() error {
	log.Printf("Task %d initializing.\n", this.id)
	this.waiter.Add(1)
	return nil
}
func (this *SimpleTask) Listen() {
	log.Printf("Task %d listening.\n", this.id)
	this.waiter.Wait()
	log.Printf("Task %d completed.\n", this.id)
}
func (this *SimpleTask) Close() error {
	log.Printf("Task %d closing.\n", this.id)
	this.waiter.Done()
	log.Printf("Task %d marked as closed.\n", this.id)
	return nil
}
