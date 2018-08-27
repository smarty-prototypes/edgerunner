package edgerunner

import (
	"context"
	"sync"
)

type (
	ContextSignaler struct {
		mutex                      *sync.Mutex
		stopContext, reloadContext context.Context
		stop, reload               context.CancelFunc
	}
	ContextSignalReader struct{ stop, reload <-chan struct{} }
)

func NewContextSignaler() *ContextSignaler {
	return &ContextSignaler{mutex: &sync.Mutex{}}
}

func (this *ContextSignaler) Start() (SignalReader, bool) {
	this.mutex.Lock()
	this.mutex.Unlock()

	if this.stop != nil {
		return this.newReader(), false
	}

	this.stopContext, this.stop = context.WithCancel(context.Background())
	this.reloadContext, this.reload = context.WithCancel(this.stopContext)
	return this.newReader(), true
}
func (this *ContextSignaler) newReader() *ContextSignalReader {
	return &ContextSignalReader{stop: this.stopContext.Done(), reload: this.reloadContext.Done()}
}
func (this *ContextSignaler) Stop() {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	if this.stop != nil {
		this.stop()
		this.stop = nil
		this.reload = nil
	}
}
func (this *ContextSignaler) Reload() {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	if this.reload != nil {
		this.reload()
	}
}
func (this *ContextSignalReader) Read() bool {
	select {
	case <-this.stop:
		return false
	case <-this.reload:
		return true
	}
}
