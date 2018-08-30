package edgerunner

import "sync"

type DefaultSignaler struct {
	signals chan interface{}
	mutex   *sync.Mutex
}

func NewSignaler() *DefaultSignaler {
	return &DefaultSignaler{mutex: &sync.Mutex{}}
}

func (this *DefaultSignaler) Start() (SignalReader, bool) {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	if this.signals == nil {
		this.signals = make(chan interface{}, 2) // buffered channel
		return DefaultSignalReader{channel: this.signals}, true
	} else {
		return DefaultSignalReader{channel: this.signals}, false
	}
}
func (this *DefaultSignaler) Stop() {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	if this.signals != nil {
		close(this.signals)
		this.signals = nil
	}
}
func (this *DefaultSignaler) Signal() bool {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	if this.signals == nil {
		return false
	}

	if len(this.signals) == 0 {
		this.signals <- nil // only send a signal if one isn't waiting
	}

	return true
}

///////////////////////////////

type DefaultSignalReader struct{ channel <-chan interface{} }

func (this DefaultSignalReader) Read() bool {
	// TODO: drain the channel completely on this read operation
	_, ok := <-this.channel
	return ok
}
