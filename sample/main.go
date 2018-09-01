package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

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

func newScheduler(reader edgerunner.SignalReader) edgerunner.Scheduler {
	return edgerunner.NewConcurrentScheduler(reader, NewSimpleTask)
}

/////////////////////////////////////////////////////////////////

var counter uint64

type SimpleTask struct {
	id     uint64
	waiter *sync.WaitGroup
}

func NewSimpleTask() edgerunner.Task {
	id := atomic.AddUint64(&counter, 1)
	return &SimpleTask{id: id, waiter: &sync.WaitGroup{}}
}

func (this *SimpleTask) Init() error {
	log.Printf("%d initializing...", this.id)
	this.waiter.Add(1)
	time.Sleep(time.Second)

	if this.id%2 == 0 {
		log.Printf("%d initialized FAILED", this.id)
		return errors.New("init failed")
	}

	log.Printf("%d initialized", this.id)
	return nil
}

func (this *SimpleTask) Listen() {
	log.Printf("%d listening...", this.id)
	this.waiter.Wait()
	log.Printf("%d listen completed", this.id)
}

func (this *SimpleTask) Close() error {
	log.Printf("%d closing...", this.id)
	time.Sleep(time.Millisecond * 500)
	this.waiter.Done()
	log.Printf("%d marked as closed", this.id)
	return nil
}
