package Bucket

import (
	"time"
)

type Token struct {
}

type Bucket struct {
	Rate int
	l    time.Duration
	b    chan Token
	t    *time.Ticker
	s    chan Token
}

func NewBucket(rate int, l time.Duration) Bucket {
	b := make(chan Token, rate)

	return Bucket{
		Rate: rate,
		l:    l,
		b:    b,
	}
}

func (lb *Bucket) Start() {
	lb.t = time.NewTicker(lb.l)
	lb.s = make(chan Token)

	lb.drain()

	go func() {
		defer close(lb.s)
		for {
			select {
			case <-lb.t.C:
				lb.drain()
			case <-lb.s:
				lb.t.Stop()
				return
			}
		}
	}()
}

func (lb *Bucket) Stop() {
	lb.s <- Token{}
	<-lb.s

	lb.fill()
}

func (lb *Bucket) IsFull() bool {
	select {
	case lb.b <- Token{}:
		return false
	default:
		return true
	}
}

func (lb *Bucket) fill() {
	for i := 0; i < lb.Rate; i++ {
		select {
		case lb.b <- Token{}:
		default:
		}
	}
}

func (lb *Bucket) drain() {
	for i := 0; i < lb.Rate; i++ {
		select {
		case <-lb.b:
		default:
		}
	}
}

// Run calls the r function when the bucket isn't full
// also add a loop size param to break execution
func (lb *Bucket) Run(r func(), ls int) {
	lb.Start()
	defer lb.Stop()
	for i := 1; i < ls; i++ {
		if lb.IsFull() {
			continue
		}
		r()
	}
}
