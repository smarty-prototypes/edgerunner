package edgerunner

import (
	"log"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
)

type DefaultSignalWatcher struct {
	runner    Runner
	signals   chan os.Signal
	actions   map[os.Signal]bool
	listening int32
}

func NewSignalWatcher(runner Runner) *DefaultSignalWatcher {
	return NewDefaultSignalWatcher(runner, DefaultReloadSignals, DefaultCloseSignals)
}
func NewDefaultSignalWatcher(runner Runner, reloadSignals, closeSignals []os.Signal) *DefaultSignalWatcher {
	actions := make(map[os.Signal]bool, len(reloadSignals)+len(closeSignals))
	for _, item := range reloadSignals {
		actions[item] = false
	}
	for _, item := range closeSignals {
		actions[item] = true
	}
	return &DefaultSignalWatcher{runner: runner, actions: actions}
}

func (this *DefaultSignalWatcher) Listen() {
	if this.alreadyListening() {
		return
	}

	defer this.listeningComplete() // 2nd defer
	defer this.unsubscribe()       // 1st defer
	this.subscribe()
	this.listen()
}
func (this *DefaultSignalWatcher) listen() {
	for item := range this.signals {
		log.Println("\nSignal received:", item)
		if this.actions[item] {
			this.runner.Stop()
			break
		} else {
			this.runner.Reload()
		}
	}
}

func (this *DefaultSignalWatcher) alreadyListening() bool {
	return !atomic.CompareAndSwapInt32(&this.listening, 0, 1)
}
func (this *DefaultSignalWatcher) listeningComplete() {
	atomic.StoreInt32(&this.listening, 0)
}

func (this *DefaultSignalWatcher) subscribe() {
	allSignals := make([]os.Signal, 0, len(this.actions))
	for item := range this.actions {
		allSignals = append(allSignals, item)
	}

	this.signals = make(chan os.Signal, 16)
	signal.Notify(this.signals, allSignals...)
}
func (this *DefaultSignalWatcher) unsubscribe() {
	signal.Stop(this.signals)
	close(this.signals)
}

var DefaultReloadSignals = []os.Signal{syscall.SIGHUP}
var DefaultCloseSignals = []os.Signal{syscall.SIGINT, syscall.SIGTERM, os.Interrupt}
