package timerwheel

import (
  "errors"
  "time"
)

type Timer struct {
  wheel *TimerWheel
}

func NewTimer(interval time.Duration, count int) *Timer {
  w := NewTimerWheel(interval, count)
  if w == nil {
    return nil
  }

  return &Timer{
    wheel: w,
  }
}

func (t *Timer) Start() {
  if t.wheel != nil {
    t.wheel.Start()
  }
}

func (t *Timer) Stop() {
  if t.wheel != nil {
    t.wheel.Stop()
  }
}

func (t *Timer) AddTimer(callback func(interface{}) error, delay time.Duration, repeated bool) (id interface{}, err error) {
  if t.wheel == nil {
    err = errors.New("invalid timer handle")
    return
  }
  t.wheel.AddTask(callback, delay, repeated)
  return
}

func (t *Timer) RemoveTask(id string) error {
  if t.wheel == nil {
    return errors.New("invalid handle")
  }

  return t.wheel.RemoveTask(id)
}
