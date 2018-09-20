package edgerunner

import "sync"

type DefaultSignaler struct {
	signals chan struct{}
	mutex   *sync.Mutex
}

func NewSignaler() *DefaultSignaler {
	return &DefaultSignaler{mutex: &sync.Mutex{}}
}

func (this *DefaultSignaler) Start() (SignalReader, bool) {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	if this.signals == nil {
		this.signals = make(chan struct{}, 2) // buffered channel
		return DefaultSignalReader{channel: this.signals}, true
	} else {
		return nil, false
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
		this.signals <- struct{}{} // only send a signal if one isn't waiting
	}

	return true
}

///////////////////////////////

type DefaultSignalReader struct{ channel <-chan struct{} }

func (this DefaultSignalReader) Read() bool {
	_, channelStillOpen := <-this.channel
	return channelStillOpen // TODO: drain the channel completely on this read operation
}
