package main

import (
	"errors"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/smartystreets-prototypes/edgerunner"
)

func main() {
	runner := edgerunner.Configure(NewSimpleTask, edgerunner.Concurrent())
	runner.Start()
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
	time.Sleep(time.Second / 2)

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
	time.Sleep(time.Second / 2)
	this.waiter.Done()
	log.Printf("%d marked as closed", this.id)
	return nil
}
