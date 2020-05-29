package util

type Mutex struct {
	c chan struct{}
}

func NewMutex() *Mutex {
	c := make(chan struct{}, 1)
	return &Mutex{c: c}
}

func (this *Mutex) Lock() {
	this.c <- struct{}{}
}

func (this *Mutex) Unlock() {
	<-this.c
}

func (this *Mutex) TryLock() bool {
	select {
	case this.c <- struct{}{}:
		return true
	default:
		return false
	}
}
