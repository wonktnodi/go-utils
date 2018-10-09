package timerwheel

import (
  "fmt"
  "github.com/stretchr/testify/assert"
  "sync/atomic"
  "testing"
  "time"
)

var timer_hits uint64 = 0

type MyTask struct {
  TaskBase
}

func (tk *MyTask) Expire() {
  fmt.Println("-----------------> test Task executed")
  atomic.AddUint64(&timer_hits, 1)
}

func (tk *MyTask) SetID(v string) {
  tk.TaskBase.SetID(v)
}

func (tk *MyTask) GetID() string {
  return tk.TaskBase.GetId()
}

func (tk *MyTask) Delay() time.Duration {
  return tk.TaskBase.Delay()
}

func (tk *MyTask) SetDelay(delay time.Duration) {
  tk.TaskBase.SetDelay(delay)
}


func TestWheelAddTimer(t *testing.T) {
  //m := sync.Map{}
  //
  //v, ok := m.Load(1)

  var wheel *TimerWheel
  wheel = NewTimerWheel(time.Microsecond*200, 5)

  //wheel.AddTask(&MyTask{}, 4*time.Second)
  //wheel.AddTask(&MyTask{}, time.Microsecond*200)
  //wheel.AddTask(&MyTask{}, time.Microsecond*400)
  //wheel.AddTask(&MyTask{}, time.Microsecond*600)
  //wheel.AddTask(&MyTask{}, time.Microsecond*800)
  //
  //wheel.AddTask(&MyTask{}, time.Microsecond*1200)
  //wheel.AddTask(&MyTask{}, time.Microsecond*1400)
  //wheel.AddTask(&MyTask{}, time.Microsecond*1600)
  //wheel.AddTask(&MyTask{}, time.Microsecond*1800)
  //wheel.AddTask(&MyTask{}, time.Microsecond*2000)
  //
  //wheel.AddTask(&MyTask{}, time.Microsecond*1210)
  //wheel.AddTask(&MyTask{}, time.Microsecond*1410)
  //wheel.AddTask(&MyTask{}, time.Microsecond*1610)
  //wheel.AddTask(&MyTask{}, time.Microsecond*1810)
  //wheel.AddTask(&MyTask{}, time.Microsecond*2010)
  //
  //assert.Equal(t, 3, len(wheel.wheel[0].items), "0 ms ticker count")
  //assert.Equal(t, 3, len(wheel.wheel[1].items), "200 ms ticker count")
  //assert.Equal(t, 3, len(wheel.wheel[2].items), "400 ms ticker count")
  //assert.Equal(t, 3, len(wheel.wheel[3].items), "600 ms ticker count")
  //assert.Equal(t, 3, len(wheel.wheel[4].items), "800 ms ticker count")
}

func TestWheelTimerRemove(t *testing.T) {
  //timer_hits = 0
  //var wheel *TimerWheel
  //wheel = NewTimerWheel(time.Microsecond*200, 5)
  //tid, _ := wheel.AddTask(&MyTask{}, 3*time.Second)
  //wheel.Start()
  //time.Sleep(time.Second)
  //wheel.RemoveTask(tid)
  //time.Sleep(3 * time.Millisecond)
  //
  //opsFinal := atomic.LoadUint64(&timer_hits)
  //assert.Equal(t, uint64(0), opsFinal, "doesn't hit the timer")
  //wheel.Stop()
}

func TestWheelTimerRound(t *testing.T) {
  //var wheel *TimerWheel
  //wheel = NewTimerWheel(time.Microsecond*200, 5)
  //
  //id, _ := wheel.AddTask(&MyTask{}, 4*time.Second)
  //assert.Equal(t, 3999, wheel.wheel[0].items[id].rounds)
  //
  //id, _ = wheel.AddTask(&MyTask{}, time.Microsecond*1201)
  //assert.Equal(t, 1, wheel.wheel[2].items[id].rounds)
  //
  //id, _ = wheel.AddTask(&MyTask{}, time.Microsecond*600)
  //assert.Equal(t, 0, wheel.wheel[3].items[id].rounds)
  //
  //id, _ = wheel.AddTask(&MyTask{}, time.Microsecond*801)
  //assert.Equal(t, 1, wheel.wheel[0].items[id].rounds)
  //
  //id, _ = wheel.AddTask(&MyTask{}, time.Microsecond*701)
  //assert.Equal(t, 0, wheel.wheel[4].items[id].rounds)
  //
  //id, _ = wheel.AddTask(&MyTask{}, time.Microsecond*100)
  //assert.Equal(t, 0, wheel.wheel[1].items[id].rounds)
}

func TestWheelTimerTimeout(t *testing.T) {
  //timer_hits = 0
  //var wheel *TimerWheel
  //wheel = NewTimerWheel(time.Microsecond*200, 5)
  //_, err := wheel.AddTask(&MyTask{}, 3*time.Second)
  //assert.Nil(t, err, "add task returns")
  //wheel.Start()
  //
  //time.Sleep(4 * time.Millisecond)
  //opsFinal := atomic.LoadUint64(&timer_hits)
  //assert.Equal(t, uint64(0), opsFinal, "doesn't hit the timer")
  //
  //time.Sleep(4 * time.Second)
  //opsFinal = atomic.LoadUint64(&timer_hits)
  //assert.Equal(t, uint64(1), opsFinal, "doesn't hit the timer")
  //
  //wheel.Stop()
}

