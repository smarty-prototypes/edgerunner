package main

import "sync"

type ChannelSignaler struct {
	signals chan interface{}
	mutex   *sync.Mutex
}

func NewChannelSignaler() *ChannelSignaler {
	return &ChannelSignaler{mutex: &sync.Mutex{}}
}

func (this *ChannelSignaler) Start() SignalReader {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	if this.signals == nil {
		this.signals = make(chan interface{}, 2)
	}

	return ChannelSignalReader{channel: this.signals}
}
func (this *ChannelSignaler) Stop() {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	if this.signals != nil {
		close(this.signals)
		this.signals = nil
	}
}
func (this *ChannelSignaler) Signal() bool {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	if this.signals == nil {
		return false
	}

	if len(this.signals) > 0{
		return true // act like we received a signal but we really didn't
	}

	this.signals <- nil
	return true
}

///////////////////////////////

type ChannelSignalReader struct {
	channel <-chan interface{}
}

func (this ChannelSignalReader) Read() bool {
	_, ok := <-this.channel
	return ok
}
