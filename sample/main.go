package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"

	"github.com/smartystreets/edgerunner"
)

func main() {
	signaler := edgerunner.NewContextSignaler(context.Background())

	runner := edgerunner.NewRunner(signaler, newScheduler)

	go func() {
		signals := make(chan os.Signal, 16)
		signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
		fmt.Println("\nSignal received:", <-signals)
		runner.Stop()
	}()

	go func() {
		signals := make(chan os.Signal, 16)
		signal.Notify(signals, syscall.SIGHUP)
		for message := range signals {
			fmt.Println("\nSignal received:", message)
			runner.Reload()
		}
	}()

	runner.Start()
}

func newScheduler(reader edgerunner.SignalReader) edgerunner.Scheduler {
	return edgerunner.NewSerialScheduler(reader, NewWebTask)
}

/////////////////////////////////////////////////////////////////

type WebTask struct {
	inner    edgerunner.Task
	output   chan int8
	server   *http.Server
	listener net.Listener
	counter  int32
}

func NewWebTask() edgerunner.Task {
	this := &WebTask{
		output: make(chan int8, 16),
	}

	this.inner = NewSanitizeTask(this.output)
	this.server = &http.Server{Handler: this}

	return this
}
func (this *WebTask) Init() error {
	this.inner.Init()
	log.Println("Web Init")

	config := &net.ListenConfig{Control: reusePort}

	// this is where we take control of the port
	if listener, err := config.Listen(context.Background(), "tcp", "127.0.0.1:8080"); err != nil {
		return err
	} else {
		this.listener = listener
	}

	return nil
}

func reusePort(network, address string, conn syscall.RawConn) error {
	return conn.Control(func(descriptor uintptr) {
		syscall.SetsockoptInt(int(descriptor), syscall.SOL_SOCKET, syscall.SO_REUSEPORT, 1)
	})
}

func (this *WebTask) Listen() {
	go this.listen()
	this.inner.Listen()
}
func (this *WebTask) listen() {
	log.Println("Web Listening...")

	this.server.Serve(this.listener) // this should block
	this.server.Shutdown(context.Background())
	close(this.output)
	this.listener.Close()

	log.Println("Web Listen completed")
}

func (this *WebTask) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	value := int8(atomic.AddInt32(&this.counter, 1))
	log.Println("Web Listen", value)
	this.output <- value
}

func (this *WebTask) Close() error {
	log.Println("Web Closing")
	this.server.Close()
	log.Println("Web marked server as closed")
	return nil
}

///////////////////////////////////////////////////////////////////

type SanitizeTask struct {
	inner  edgerunner.Task
	input  chan int8
	output chan int16
}

func NewSanitizeTask(input chan int8) edgerunner.Task {
	output := make(chan int16, 16)
	return &SanitizeTask{
		inner:  NewVerifyTask(output),
		input:  input,
		output: output,
	}
}
func (this *SanitizeTask) Init() error {
	this.inner.Init()
	log.Println("Sanitize Init")
	return nil
}
func (this *SanitizeTask) Listen() {
	go this.listen()    // start my listener in the background
	this.inner.Listen() // wait for inner listener to complete before we exit

}
func (this *SanitizeTask) listen() {
	log.Println("Sanitize Listening...")
	for item := range this.input {
		log.Println("Sanitize:", item)
		this.output <- int16(item)
	}

	close(this.output)
	log.Println("Sanitize Listen completed")
}

func (this *SanitizeTask) Close() error {
	return nil
}

///////////////////////////////////////////////////////////////////

type VerifyTask struct {
	inner  edgerunner.Task
	input  chan int16
	output chan int32
}

func NewVerifyTask(input chan int16) edgerunner.Task {
	output := make(chan int32, 16)
	return &VerifyTask{
		inner:  NewLogTask(output),
		input:  input,
		output: output,
	}
}
func (this *VerifyTask) Init() error {
	this.inner.Init()
	log.Println("Verify Init")
	return nil
}
func (this *VerifyTask) Listen() {
	go this.listen()    // start my listener in the background
	this.inner.Listen() // wait for inner listener to complete before we exit

}
func (this *VerifyTask) listen() {
	log.Println("Verify Listening...")

	for item := range this.input {
		log.Println("Verify:", item)
		this.output <- int32(item)
	}

	close(this.output)
	log.Println("Verify Listen completed")
}

func (this *VerifyTask) Close() error {
	return nil
}

///////////////////////////////////////////////////////////////////

type LogTask struct {
	input chan int32
}

func NewLogTask(input chan int32) edgerunner.Task {
	return &LogTask{input: input}
}
func (this *LogTask) Init() error {
	log.Println("Log Init")
	return nil
}
func (this *LogTask) Listen() {
	log.Println("Log Listening...")

	for item := range this.input {
		log.Println("Log:", item)
	}

	log.Println("Log Listen completed")
}
func (this *LogTask) Close() error {
	return nil
}

///////////////////////////////////////////////////////////////////
