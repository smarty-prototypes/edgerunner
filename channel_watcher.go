package main

type ChannelWatcher struct {
	channel <-chan interface{}
}

func NewChannelWatcher(channel <-chan interface{}) ChannelWatcher {
	return ChannelWatcher{channel: channel}
}

func (this ChannelWatcher) Read() bool {
	_, ok := <-this.channel
	return ok
}
