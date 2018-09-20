package edgerunner

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type SignalRunner struct {
	inner   Runner
	mutex   *sync.Mutex
	signals chan os.Signal
	reload  map[os.Signal]bool
	all     []os.Signal
}

func NewSignalRunner(inner Runner, reload, terminate []os.Signal) *SignalRunner {
	return &SignalRunner{
		inner:  inner,
		mutex:  &sync.Mutex{},
		reload: sliceToMap(reload),
		all:    append(reload, terminate...),
	}
}
func sliceToMap(slice []os.Signal) map[os.Signal]bool {
	output := make(map[os.Signal]bool, len(slice))

	for _, item := range slice {
		output[item] = true
	}

	return output
}

func (this *SignalRunner) Start() error {
	if !this.start() {
		return nil
	}

	return this.inner.Start()
}
func (this *SignalRunner) Stop() {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	if this.signals == nil {
		return
	}

	signal.Stop(this.signals)
	close(this.signals)
	this.signals = nil
	this.inner.Stop()
}
func (this *SignalRunner) Reload() bool {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	return this.signals != nil && this.inner.Reload()
}

func (this *SignalRunner) start() bool {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	if this.signals != nil {
		return false // already started
	}

	this.signals = make(chan os.Signal, 16)
	signal.Notify(this.signals, this.all...)
	go this.listen(this.signals)
	return true
}
func (this *SignalRunner) listen(input <-chan os.Signal) {
	defer this.Stop()

	for item := range input {
		fmt.Println("\nSignal received:", item)
		if this.reload[item] {
			this.Reload()
		} else {
			break
		}
	}
}

var DefaultReloadSignals = []os.Signal{syscall.SIGHUP}
var DefaultShutdownSignals = []os.Signal{syscall.SIGINT, syscall.SIGTERM}
