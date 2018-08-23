package main

type Task interface {
	// prepares the instance to handle traffic
	Init() error

	// blocking and can be called at most once (when listen exits, the runnable instance is considered closed)
	Listen()

	// close is non-blocking and concurrent (meaning it can be called during initialize or listen)
	Close() error
}
