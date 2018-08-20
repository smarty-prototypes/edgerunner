package main

import "io"

type Runnable interface {
	// prepares the instance to handle traffic
	Initialize() error

	// blocking and can be called at most once (when listen exits, the runnable instance is considered closed)
	Listen()

	// close is non-blocking and concurrent (meaning it can be called during initialize or listen)
	io.Closer

	// can more than one instance run simultaneously and to what degree
	Concurrency() int
}

const (
	// multiple instances can be listening at the same time
	ConcurrencyFull int = iota

	// multiple instances can be initialized but only one can listen at a time
	ConcurrencyPartial

	// only one instance can be initialized and listen at the same time
	ConcurrencyNone
)