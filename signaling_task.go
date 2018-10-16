package edgerunner

type SignalingTask struct {
	inner      Task
	initialize chan<- error
	listen     chan<- bool
	closer     chan struct{}
}

func NewSignalingTask(inner Task, initialize chan<- error, listen chan<- bool) *SignalingTask {
	return &SignalingTask{
		inner:      inner,
		initialize: initialize,
		listen:     listen,
		closer:     make(chan struct{}, 2),
	}
}

func (this *SignalingTask) Init() error {
	err := this.inner.Init()
	this.initialize <- err
	return err
}
func (this *SignalingTask) Listen() {
	this.inner.Listen()
	select {
	case _, open := <-this.closer:
		this.listen <- !open // only succeed if closed properly
	}
}
func (this *SignalingTask) Close() error {
	close(this.closer)
	return this.inner.Close()
}
