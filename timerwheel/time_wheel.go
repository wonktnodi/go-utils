package timerwheel

import (
  "crypto/md5"
  "crypto/rand"
  "encoding/base64"
  "encoding/hex"
  "errors"
  "io"
  "sync"
  "time"
)

const (
  Default_tick_duration = time.Second // default tick duration 100ms
  Default_wheel_count   = 512         // default slot count
)

type TimerWheel struct {
  state         int           // start(1)or stop(-1) status
  tickDuration  time.Duration // duration timer wheels
  roundDuration time.Duration // one round duration
  wheelCount    int           // slot count
  wheel         []*Iterator   // slots
  tick          *time.Ticker  // ticker
  lock          sync.Mutex    // lock
  wheelCursor   int           // current slot index
  mask          int           // max index of slot
}

//Iterator for time wheel slot
type Iterator struct {
  items map[string]*WheelTimeOut
}

//Timeout object
type WheelTimeOut struct {
  delay    time.Duration // Delay time
  repeated bool
  index    int // slot index
  rounds   int // round times
  id       string
  task     func(interface{}) error
}

func NewTimerWheel(duration time.Duration, count int) *TimerWheel {
  return &TimerWheel{
    tickDuration:  duration,
    wheelCount:    count,
    wheel:         createWheel(count),
    wheelCursor:   0,
    mask:          count - 1,
    roundDuration: duration * time.Duration(count),
  }
}

func (t *TimerWheel) Start() {
  t.lock.Lock()
  t.tick = time.NewTicker(t.tickDuration)
  defer t.lock.Unlock()
  go func() {
    for {
      select {
      case <-t.tick.C:
        t.wheelCursor++
        if t.wheelCursor == t.wheelCount {
          t.wheelCursor = 0
        }
        //check the timeout event in current slot
        iterator := t.wheel[t.wheelCursor]
        tasks := t.fetchExpiredTimeouts(iterator)
        t.notifyExpiredTimeOut(tasks)
      }
    }
  }()
}

func (t *TimerWheel) Stop() {
  t.tick.Stop()
}

func createWheel(count int) []*Iterator {
  arr := make([]*Iterator, count)

  for v := 0; v < count; v++ {
    arr[v] = &Iterator{items: make(map[string]*WheelTimeOut)}
  }
  return arr
}

type timerEntry struct {
  id       string
  delay    time.Duration
  callback func(id interface{}) error
}

func (t *TimerWheel) AddTask(task func(interface{}) error, delay time.Duration, repeated bool) (string, error) {
  if task == nil {
    return "", errors.New("task is empty")
  }
  if delay <= 0 {
    return "", errors.New("Delay Must be greater than zero ")
  }
  timeOut := &WheelTimeOut{
    delay:    delay,
    repeated: repeated,
    task:     task,
  }

  tid, err := t.scheduleTimeOut(timeOut)

  return tid, err
}

func (t *TimerWheel) RemoveTask(taskId string) error {
  for _, it := range t.wheel {
    for k, _ := range it.items {
      if taskId == k {
        delete(it.items, k)
      }
    }
  }
  return nil
}

func (t *TimerWheel) scheduleTimeOut(timeOut *WheelTimeOut) (string, error) {
  if timeOut.delay < t.tickDuration {
    timeOut.delay = t.tickDuration
  }
  lastRoundDelay := timeOut.delay % t.roundDuration
  lastTickDelay := timeOut.delay % t.tickDuration

  //calculate slot index
  relativeIndex := lastRoundDelay / t.tickDuration
  if lastTickDelay != 0 {
    relativeIndex = relativeIndex + 1
  }
  //calculate round count
  remainingRounds := timeOut.delay / t.roundDuration
  if lastRoundDelay == 0 {
    remainingRounds = remainingRounds - 1
  }
  t.lock.Lock()
  defer t.lock.Unlock()
  stopIndex := t.wheelCursor + int(relativeIndex)
  if stopIndex >= t.wheelCount {
    stopIndex = stopIndex - t.wheelCount
    if remainingRounds > 0 {
      timeOut.rounds = int(remainingRounds) + 1
    }
  } else {
    timeOut.rounds = int(remainingRounds)
  }
  timeOut.index = stopIndex
  item := t.wheel[stopIndex]
  if item == nil {
    item = &Iterator{
      items: make(map[string]*WheelTimeOut),
    }
  }

  key, err := GetGuid()
  if err != nil {
    return "", err
  }
  item.items[key] = timeOut
  t.wheel[stopIndex] = item
  timeOut.id = key
  //log.Tracef("timeout[%s] item: %d, round", timeOut.id, timeOut.index, timeOut.rounds)
  return key, nil
}

func (t *TimerWheel) fetchExpiredTimeouts(iterator *Iterator) []*WheelTimeOut {
  t.lock.Lock()
  defer t.lock.Unlock()

  task := []*WheelTimeOut{}

  for k, v := range iterator.items {
    if v.rounds <= 0 { // expired
      task = append(task, v)
      delete(iterator.items, k)
    } else {
      v.rounds--
    }
  }

  return task
}

func (t *TimerWheel) notifyExpiredTimeOut(tasks []*WheelTimeOut) {
  for _, task := range tasks {
    item := task
    go func() {
      item.task(item.id)
      if item.repeated {
        t.AddTask(item.task, item.delay, item.repeated)
        //t.scheduleTimeOut(item)
      }
    }()
  }
}

func GetMd5String(s string) string {
  h := md5.New()
  h.Write([]byte(s))
  return hex.EncodeToString(h.Sum(nil))

}

//return guid str
func GetGuid() (string, error) {

  b := make([]byte, 48)
  if _, err := io.ReadFull(rand.Reader, b); err != nil {
    return "", err
  }

  return GetMd5String(base64.URLEncoding.EncodeToString(b)), nil
}
