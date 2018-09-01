package edgerunner

import (
	"io"
	"sync"
)

type ConcurrentScheduler struct {
	reader  SignalReader
	factory TaskFactory
	waiter  *sync.WaitGroup
	mutex   *sync.Mutex
	err     error
	active  []io.Closer
}

func NewConcurrentScheduler(reader SignalReader, factory TaskFactory) *ConcurrentScheduler {
	return &ConcurrentScheduler{
		reader:  reader,
		factory: factory,
		waiter:  &sync.WaitGroup{},
		mutex:   &sync.Mutex{},
	}
}

func (this *ConcurrentScheduler) Schedule() error {
	for this.schedule() {
	}

	this.closeAll()
	this.waiter.Wait()
	return this.err
}
func (this *ConcurrentScheduler) schedule() bool {
	go this.runTask(this.newTask())
	return this.reader.Read()
}
func (this *ConcurrentScheduler) newTask() (io.Closer, Task) {
	this.waiter.Add(1)
	item := this.factory()

	this.mutex.Lock()
	defer this.mutex.Unlock()

	this.active = append(this.active, item)
	if len(this.active) > 1 {
		return this.active[0], item
	} else {
		return nil, item
	}
}
func (this *ConcurrentScheduler) runTask(previous io.Closer, current Task) {
	defer this.waiter.Done()

	if this.storeError(current.Init()) {
		this.closeTask(current) // current one failed to start, mark it as closed
	} else {
		this.closeTask(previous)
		current.Listen()
	}
}

func (this *ConcurrentScheduler) closeTask(task io.Closer) {
	if task == nil {
		return
	}

	go task.Close() // mark as closed in the background

	this.mutex.Lock()
	defer this.mutex.Unlock()

	for i, item := range this.active {
		if item == task {
			this.active = append(this.active[:i], this.active[i+1:]...)
		}
	}
}
func (this *ConcurrentScheduler) closeAll() {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	for _, item := range this.active {
		item.Close()
	}

	this.active = nil
}

func (this *ConcurrentScheduler) storeError(err error) bool {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	this.err = err
	return err != nil
}
