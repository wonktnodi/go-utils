package log

import (
  "fmt"
  "log"
  "os"
  "path"
  "runtime"
  "strings"
  "sync/atomic"
  "time"
)

// LogLevel is the log level type.
type LogLevel int

const (
  Lfile = 1 << 6 // file name
  Lfunc = 1 << 7 // func name
  Lline = 1 << 8 // line number
)

const (
  TRACE LogLevel = iota
  // DEBUG represents debug log level.
  DEBUG
  // INFO represents info log level.
  INFO
  // WARN represents warn log level.
  WARN
  // ERROR represents error log level.
  ERROR
  // FATAL represents fatal log level.
  FATAL
)

var (
  started        int32
  loggerInstance *LogAdaptor
  logger         Logger
  tagName        = map[LogLevel]string{
    TRACE: "TRC",
    DEBUG: "DBG",
    INFO:  "INF",
    WARN:  "WRN",
    ERROR: "ERR",
    FATAL: "FTL",
  }
)

func NewLogInstance(decorators ...func(Logger) Logger) Logger {
  inst := Logger{}
  for _, decorator := range decorators {
    inst = decorator(inst)
  }
  var logger *log.Logger
  var segment *logSegment
  if inst.logPath != "" {
    segment = newLogSegment(inst.unit, inst.logPath, inst.name)
  }
  if segment != nil {
    logger = log.New(segment, "", log.LstdFlags|log.Lmicroseconds)
    inst.segment = segment
  } else {
    logger = log.New(os.Stderr, "", log.LstdFlags|log.Lmicroseconds)
  }
  inst.logger = logger
  return inst
}

// Start returns a decorated innerLogger.
func Start(decorators ...func(Logger) Logger) *LogAdaptor {
  if atomic.CompareAndSwapInt32(&started, 0, 1) {
    logger = NewLogInstance(decorators...)
    loggerInstance = NewAdaptorFromInstance(&logger, 4)
    return loggerInstance
  }
  //return nil
  panic("Start() already called")
}

func (l Logger) Release() {
  if l.printStack {
    traceInfo := make([]byte, 1<<16)
    n := runtime.Stack(traceInfo, true)
    l.logger.Printf("%s", traceInfo[:n])
    if l.isStdout {
      log.Printf("%s", traceInfo[:n])
    }
  }
  if l.segment != nil {
    l.segment.Close()
  }
  l.segment = nil
  l.logger = nil
}

func Stop() {
  logger.Stop()
}

// Stop stops the logger.
func (l Logger) Stop() {
  if atomic.CompareAndSwapInt32(&l.stopped, 0, 1) {
    l.Release()
    atomic.StoreInt32(&started, 0)
  }
}

// logSegment implements io.Writer
type logSegment struct {
  unit         time.Duration
  logPath      string
  logFileName  string
  logFile      *os.File
  pid          int
  timeToCreate <-chan time.Time
}

func newLogSegment(unit time.Duration, logPath string, fileName string) *logSegment {
  now := time.Now()
  if logPath != "" {
    err := os.MkdirAll(logPath, os.ModePerm)
    if err != nil {
      fmt.Fprintln(os.Stderr, err)
      return nil
    }
    name := strings.TrimSpace(fileName)
    if name == "" {
      name = getLogName()
    }
    filename := path.Join(logPath, name)
    logFile, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
    if err != nil {
      if os.IsNotExist(err) {
        logFile, err = os.Create(path.Join(logPath, name))
        if err != nil {
          fmt.Fprintln(os.Stderr, err)
          return nil
        }
      } else {
        fmt.Fprintln(os.Stderr, err)
        return nil
      }
    }
    next := now.Truncate(unit).Add(unit)
    var timeToCreate <-chan time.Time
    if unit == time.Hour || unit == time.Minute {
      timeToCreate = time.After(next.Sub(time.Now()))
    }
    return &logSegment{
      unit:         unit,
      logPath:      logPath,
      logFileName:  filename,
      logFile:      logFile,
      pid:          os.Getpid(),
      timeToCreate: timeToCreate,
    }
  }
  return nil
}

func (ls *logSegment) Write(p []byte) (n int, err error) {
  if ls.timeToCreate != nil && ls.logFile != os.Stdout && ls.logFile != os.Stderr {
    select {
    case current := <-ls.timeToCreate:
      backup := getLogFileName(current)
      val := path.Join(ls.logPath, backup)
      ls.logFile.Close()
      ls.logFile = nil
      os.Rename(ls.logFileName, val)
      
      ls.logFile, err = os.Create(ls.logFileName)
      if err != nil {
        // log into stderr if we can't create new file
        fmt.Fprintln(os.Stderr, err)
        ls.logFile = os.Stderr
      } else {
        next := current.Truncate(ls.unit).Add(ls.unit)
        ls.timeToCreate = time.After(next.Sub(time.Now()))
      }
    default:
      // do nothing
    }
  }
  return ls.logFile.Write(p)
}

func (ls *logSegment) Close() {
  ls.logFile.Close()
}

func getLogName() string {
  return path.Base(os.Args[0]) + ".log"
}

func getLogFileName(t time.Time) string {
  proc := path.Base(os.Args[0])
  now := time.Now()
  year := now.Year()
  month := now.Month()
  day := now.Day()
  hour := now.Hour()
  minute := now.Minute()
  pid := os.Getpid()
  return fmt.Sprintf("%s.%04d-%02d-%02d-%02d-%02d.%d.log",
    proc, year, month, day, hour, minute, pid)
}

type LogAdaptor struct {
  logger    *Logger
  calldepth int
}

func NewAdaptorFromInstance(log *Logger, callDepth int) *LogAdaptor {
  return &LogAdaptor{
    logger:    log,
    calldepth: callDepth,
  }
}

func NewAdaptor(callDepth int) *LogAdaptor {
  return &LogAdaptor{
    logger:    &logger,
    calldepth: callDepth,
  }
}

func (l *LogAdaptor) Print(v ...interface{}) {
  l.logger.Print(v...)
}

func (l *LogAdaptor) Printf(format string, v ...interface{}) {
  l.logger.Printf(format, v...)
}

func (l *LogAdaptor) Tracef(format string, v ...interface{}) {
  l.logger.doPrintfN(l.calldepth, TRACE, format, v...)
}

// Debugf prints formatted debug log.
func (l *LogAdaptor) Debugf(format string, v ...interface{}) {
  l.logger.doPrintfN(l.calldepth, DEBUG, format, v...)
}

// Infof prints formatted info log.
func (l *LogAdaptor) Infof(format string, v ...interface{}) {
  l.logger.doPrintfN(l.calldepth, INFO, format, v...)
}

// Warnf prints formatted warn log.
func (l *LogAdaptor) Warnf(format string, v ...interface{}) {
  l.logger.doPrintfN(l.calldepth, WARN, format, v...)
}

// Errorf prints formatted error log.
func (l *LogAdaptor) Errorf(format string, v ...interface{}) {
  l.logger.doPrintfN(l.calldepth, ERROR, format, v...)
}

// Fatalf prints formatted fatal log and exits.
func (l *LogAdaptor) Fatalf(format string, v ...interface{}) {
  l.logger.doPrintfN(l.calldepth, FATAL, format, v...)
  os.Exit(1)
}

// Traceln prints debug log.
func (l *LogAdaptor) Traceln(v ...interface{}) {
  l.logger.doPrintlnN(l.calldepth, TRACE, v...)
}

// Debugln prints debug log.
func (l *LogAdaptor) Debugln(v ...interface{}) {
  l.logger.doPrintlnN(l.calldepth, DEBUG, v...)
}

// Infoln prints info log.
func (l *LogAdaptor) Infoln(v ...interface{}) {
  l.logger.doPrintlnN(l.calldepth, INFO, v...)
}

// Warnln prints warn log.
func (l *LogAdaptor) Warnln(v ...interface{}) {
  l.logger.doPrintlnN(l.calldepth, WARN, v...)
}

// Errorln prints error log.
func (l *LogAdaptor) Errorln(v ...interface{}) {
  l.logger.doPrintlnN(l.calldepth, ERROR, v...)
}

// Fatalln prints fatal log and exits.
func (l *LogAdaptor) Fatalln(v ...interface{}) {
  l.logger.doPrintlnN(l.calldepth, FATAL, v...)
  os.Exit(1)
}

func (l *LogAdaptor) Write(p []byte) (n int, err error) {
  l.logger.WriteN(l.calldepth, p)
  return
}

func (l *LogAdaptor) SetCallDepth(callDepth int) {
  l.calldepth = callDepth
}

func (l *LogAdaptor) SetLevel(level LogLevel) {
  SetLevel(l.logger, level)
}

// Logger is the logger type.
type Logger struct {
  logger     *log.Logger
  level      LogLevel
  segment    *logSegment
  stopped    int32
  logPath    string
  name       string
  flags      int32
  unit       time.Duration
  isStdout   bool
  printStack bool
}

func (l Logger) Write(p []byte) (n int, err error) {
  l.doPrintln(TRACE, string(p))
  return
}

func (l Logger) WriteN(callDepth int, p []byte) (n int, err error) {
  l.doPrintlnN(callDepth, TRACE, string(p))
  return
}

func (l Logger) Print(v ...interface{}) {
  l.doPrintln(DEBUG, v...)
}

func (l Logger) PrintN(callDepth int, v ...interface{}) {
  l.doPrintlnN(callDepth, DEBUG, v...)
}

func (l Logger) Printf(format string, v ...interface{}) {
  if l.logger == nil {
    return
  }
  l.logger.Printf(format, v...)
}

func (l Logger) PrintfN(callDepth int, format string, v ...interface{}) {
  if l.logger == nil {
    return
  }
  l.logger.Output(callDepth, fmt.Sprintf(format, v...))
}

func (l Logger) doPrintfN(callDepth int, level LogLevel, format string, v ...interface{}) {
  if l.logger == nil {
    return
  }
  if level >= l.level {
    if l.flags > 0 {
      funcName, fileName, lineNum := getRuntimeInfo(callDepth)
      if l.flags&(Lfile|Lline) != 0 {
        if l.flags&Lline == 0 {
          format = fmt.Sprintf("%3s [%s] (%s): %s", tagName[level], path.Base(funcName),
            path.Base(fileName), format)
        } else {
          format = fmt.Sprintf("%3s [%s] (%s:%d): %s", tagName[level], path.Base(funcName),
            path.Base(fileName), lineNum, format)
        }
      } else {
        format = fmt.Sprintf("%3s [%s]: %s", tagName[level], path.Base(funcName), format)
      }
      
    } else {
      format = fmt.Sprintf("%3s: %s", tagName[level], format)
    }
    
    l.logger.Printf(format, v...)
    if l.isStdout {
      log.Printf(format, v...)
    }
    if level == FATAL {
      os.Exit(1)
    }
  }
}

func (l Logger) doPrintf(level LogLevel, format string, v ...interface{}) {
  l.doPrintfN(3, level, format, v...)
  //if l.logger == nil {
  //  return
  //}
  //if level >= l.level {
  //  if l.flags > 0 {
  //    funcName, fileName, lineNum := getRuntimeInfo(3)
  //    if l.flags&(Lfile|Lline) != 0 {
  //      if l.flags&Lline == 0 {
  //        format = fmt.Sprintf("%3s [%s] (%s): %s", tagName[level], path.Base(funcName),
  //          path.Base(fileName), format)
  //      } else {
  //        format = fmt.Sprintf("%3s [%s] (%s:%d): %s", tagName[level], path.Base(funcName),
  //          path.Base(fileName), lineNum, format)
  //      }
  //    } else {
  //      format = fmt.Sprintf("%3s [%s]: %s", tagName[level], path.Base(funcName), format)
  //    }
  //
  //  } else {
  //    format = fmt.Sprintf("%3s: %s", tagName[level], format)
  //  }
  //
  //  l.logger.Printf(format, v...)
  //  if l.isStdout {
  //    log.Printf(format, v...)
  //  }
  //  if level == FATAL {
  //    os.Exit(1)
  //  }
  //}
}

func (l Logger) doPrintlnN(callDepth int, level LogLevel, v ...interface{}) {
  if l.logger == nil {
    return
  }
  if level >= l.level {
    var prefix string
    if l.flags > 0 {
      funcName, fileName, lineNum := getRuntimeInfo(callDepth)
      if l.flags&(Lfile|Lline) != 0 {
        if l.flags&Lline != 0 {
          prefix = fmt.Sprintf("%3s [%s] (%s:%d): ", tagName[level], path.Base(funcName),
            path.Base(fileName), lineNum)
        } else {
          prefix = fmt.Sprintf("%3s [%s] (%s): ", tagName[level], path.Base(funcName),
            path.Base(fileName))
        }
      } else {
        prefix = fmt.Sprintf("%3s [%s]: ", tagName[level], path.Base(funcName))
      }
      
    } else {
      prefix = fmt.Sprintf("%3s: ", tagName[level])
    }
    
    value := fmt.Sprintf("%s%s", prefix, fmt.Sprintln(v...))
    l.logger.Print(value)
    if l.isStdout {
      log.Print(value)
    }
    if level == FATAL {
      os.Exit(1)
    }
  }
}

func (l Logger) doPrintln(level LogLevel, v ...interface{}) {
  l.doPrintlnN(3, level, v...)
  //if l.logger == nil {
  //  return
  //}
  //if level >= l.level {
  //  var prefix string
  //  if l.flags > 0 {
  //    funcName, fileName, lineNum := getRuntimeInfo()
  //    if l.flags&(Lfile|Lline) != 0 {
  //      if l.flags&Lline != 0 {
  //        prefix = fmt.Sprintf("%3s [%s] (%s:%d): ", tagName[level], path.Base(funcName),
  //          path.Base(fileName), lineNum)
  //      } else {
  //        prefix = fmt.Sprintf("%3s [%s] (%s): ", tagName[level], path.Base(funcName),
  //          path.Base(fileName))
  //      }
  //    } else {
  //      prefix = fmt.Sprintf("%3s [%s]: ", tagName[level], path.Base(funcName))
  //    }
  //
  //  } else {
  //    prefix = fmt.Sprintf("%3s: ", tagName[level])
  //  }
  //
  //  value := fmt.Sprintf("%s%s", prefix, fmt.Sprintln(v...))
  //  l.logger.Print(value)
  //  if l.isStdout {
  //    log.Print(value)
  //  }
  //  if level == FATAL {
  //    os.Exit(1)
  //  }
  //}
}

func getRuntimeInfo(callDepth int) (string, string, int) {
  pc, fn, ln, ok := runtime.Caller(callDepth) // 3 steps up the stack frame
  if !ok {
    fn = "???"
    ln = 0
  }
  function := "???"
  caller := runtime.FuncForPC(pc)
  if caller != nil {
    function = caller.Name()
  }
  return function, fn, ln
}

func SetLevel(l *Logger, level LogLevel) Logger {
  l.level = level
  return *l
}

// DebugLevel sets log level to debug.
func DebugLevel(l Logger) Logger {
  l.level = DEBUG
  return l
}

// InfoLevel sets log level to info.
func InfoLevel(l Logger) Logger {
  l.level = INFO
  return l
}

// WarnLevel sets log level to warn.
func WarnLevel(l Logger) Logger {
  l.level = WARN
  return l
}

// ErrorLevel sets log level to error.
func ErrorLevel(l Logger) Logger {
  l.level = ERROR
  return l
}

// FatalLevel sets log level to fatal.
func FatalLevel(l Logger) Logger {
  l.level = FATAL
  return l
}

// LogFilePath returns a function to set the log file path.
func LogFilePath(p, name string) func(Logger) Logger {
  return func(l Logger) Logger {
    l.logPath = p
    l.name = name
    return l
  }
}
func LogFlags(f int32) func(Logger) Logger {
  return func(l Logger) Logger {
    l.flags = f
    return l
  }
}

// EveryHour sets new log file created every hour.
func EveryHour(l Logger) Logger {
  l.unit = time.Hour
  return l
}

// EveryMinute sets new log file created every minute.
func EveryMinute(l Logger) Logger {
  l.unit = time.Minute
  return l
}

// AlsoStdout sets log also output to stdio.
func AlsoStdout(l Logger) Logger {
  l.isStdout = true
  return l
}

// PrintStack sets log output the stack trace info.
func PrintStack(l Logger) Logger {
  l.printStack = true
  return l
}

// Tracef prints formatted trace log.
func Tracef(format string, v ...interface{}) {
  loggerInstance.Tracef(format, v...)
}

// Debugf prints formatted debug log.
func Debugf(format string, v ...interface{}) {
  loggerInstance.Debugf(format, v...)
}

// Infof prints formatted info log.
func Infof(format string, v ...interface{}) {
  loggerInstance.Infof(format, v...)
}

// Warnf prints formatted warn log.
func Warnf(format string, v ...interface{}) {
  loggerInstance.Warnf(format, v...)
}

// Errorf prints formatted error log.
func Errorf(format string, v ...interface{}) {
  loggerInstance.Errorf(format, v...)
}

// Fatalf prints formatted fatal log and exits.
func Fatalf(format string, v ...interface{}) {
  loggerInstance.Fatalf(format, v...)
  os.Exit(1)
}

// Traceln prints debug log.
func Traceln(v ...interface{}) {
  loggerInstance.Traceln(v...)
}

// Debugln prints debug log.
func Debugln(v ...interface{}) {
  loggerInstance.Debugln(v...)
}

// Infoln prints info log.
func Infoln(v ...interface{}) {
  loggerInstance.Infoln(v...)
}

// Warnln prints warn log.
func Warnln(v ...interface{}) {
  loggerInstance.Warnln(v...)
}

// Errorln prints error log.
func Errorln(v ...interface{}) {
  loggerInstance.Errorln(v...)
}

// Fatalln prints fatal log and exits.
func Fatalln(v ...interface{}) {
  loggerInstance.Fatalln(v...)
  os.Exit(1)
}

func Write(p []byte) {
  loggerInstance.Write(p)
}

func SetCallDepth(callDepth int) {
  loggerInstance.SetCallDepth(callDepth)
}
