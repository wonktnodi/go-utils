package log

import (
	"testing"
	"time"
)

func TestStdErrLogger(t *testing.T) {
	defer Start().Stop()

	Traceln("hello tracer")
	Debugln("hello debuger")
	Infoln("Hello, Mike")
	Warnln("This might be painful but...")
	Errorln("You have to go through it until sunshine comes out")
	Infoln("Those were the days hard work forever pays")
}

// EveryMinute sets new log file created every minute.
func EverySecond(l Logger) Logger {
	l.unit = time.Second
	return l
}

func TestFileLoggerEveryMinute(t *testing.T) {
	defer Start(LogFilePath("./log"), EveryMinute).Stop()

	{
		// after one minute a new log created, uncomment it to have a try!
		//time.Sleep(time.Second*time.Duration(10))

		Infof("%s", "Jingle bells, jingle bells,")
		Warnf("%s", "Jingle all the way.")
		Errorf("%s", "Oh! what fun it is to ride")
		Infof("%s", "In a one-horse open sleigh.")
	}
	time.Sleep(time.Minute)
	{
		Traceln("hello tracer")
		Debugln("hello debuger")
		Infoln("Hello, Mike")
		Warnln("This might be painful but...")
		Errorln("You have to go through it until sunshine comes out")
		Infoln("Those were the days hard work forever pays")
	}
	time.Sleep(time.Minute)
	{
		Traceln("hello tracer")
		Debugln("hello debuger")
		Infoln("Hello, Mike")
		Warnln("This might be painful but...")
		Errorln("You have to go through it until sunshine comes out")
		Infoln("Those were the days hard work forever pays")
	}

}
