package edgerunner

import (
	"context"
	"sync"
)

type (
	ContextSignaler struct {
		mutex                *sync.Mutex
		stopCtx, reloadCtx   context.Context
		stopFunc, reloadFunc context.CancelFunc
	}
	ContextSignalReader struct{ stop, reload <-chan struct{} }
)

func NewContextSignaler() *ContextSignaler {
	return &ContextSignaler{mutex: &sync.Mutex{}}
}

func (this *ContextSignaler) Start() (SignalReader, bool) {
	this.mutex.Lock()
	this.mutex.Unlock()

	if this.stopFunc != nil {
		return this.newReader(), false
	}

	this.stopCtx, this.stopFunc = context.WithCancel(context.Background())
	this.reloadCtx, this.reloadFunc = context.WithCancel(context.Background())
	return this.newReader(), true
}
func (this *ContextSignaler) newReader() *ContextSignalReader {
	return &ContextSignalReader{stop: this.stopCtx.Done(), reload: this.reloadCtx.Done()}
}
func (this *ContextSignaler) Stop() {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	if this.stopFunc != nil {
		this.stopFunc()
		this.stopFunc = nil
		this.reloadFunc = nil
	}
}
func (this *ContextSignaler) Signal() bool {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	if this.reloadFunc == nil {
		return false
	}

	this.reloadFunc()
	return true
}
func (this *ContextSignalReader) Read() bool {
	select {
	case <-this.stop:
		return false
	case <-this.reload:
		return true
	}
}
