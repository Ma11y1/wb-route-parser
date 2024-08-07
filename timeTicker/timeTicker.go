package timeTicker

import "time"

type TimeTicker struct {
	ticker    *time.Ticker
	duration  time.Duration
	frequency time.Duration
	callback  func()
	chanStop  chan struct{}
	isStarted bool
}

func NewTimeTicker(frequency int) *TimeTicker {
	return &TimeTicker{
		ticker:    time.NewTicker(time.Duration(frequency) * time.Millisecond),
		duration:  time.Millisecond,
		frequency: time.Duration(frequency) * time.Millisecond,
		callback:  func() {},
		chanStop:  make(chan struct{}),
		isStarted: false,
	}
}

func (t *TimeTicker) SetCallback(cb func()) {
	t.callback = cb
}

func (t *TimeTicker) Start() {
	if t.isStarted {
		return
	}

	t.isStarted = true
	t.ticker.Reset(t.frequency)

	go func() {
		for {
			select {
			case <-t.ticker.C:
				t.callback()
			case <-t.chanStop:
				t.ticker.Stop()
				return
			}
		}
	}()
}

func (t *TimeTicker) Stop() {
	if t.isStarted {
		t.chanStop <- struct{}{}
		t.isStarted = false
	}
}

func (t *TimeTicker) Reset(frequency int) {
	if frequency <= 0 {
		return
	}

	t.frequency = time.Duration(frequency) * t.duration
	if t.isStarted {
		t.Stop()
		t.Start()
	}
}

func (t *TimeTicker) IsStarted() bool {
	return t.isStarted
}
