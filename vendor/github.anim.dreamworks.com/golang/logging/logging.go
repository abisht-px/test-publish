package logging

import "github.com/sirupsen/logrus"

const (
	TraceLevel = logrus.TraceLevel
	DebugLevel = logrus.DebugLevel
	InfoLevel  = logrus.InfoLevel
	WarnLevel  = logrus.WarnLevel
	ErrorLevel = logrus.ErrorLevel
	FatalLevel = logrus.FatalLevel
	PanicLevel = logrus.PanicLevel
)

var (
	log *logrus.Logger // The global Logrus instance

	Trace   func(args ...interface{})
	Traceln func(args ...interface{})
	Tracef  func(format string, args ...interface{})

	Debug   func(args ...interface{})
	Debugln func(args ...interface{})
	Debugf  func(format string, args ...interface{})

	Info   func(args ...interface{})
	Infoln func(args ...interface{})
	Infof  func(format string, args ...interface{})

	Warn   func(args ...interface{})
	Warnln func(args ...interface{})
	Warnf  func(format string, args ...interface{})

	Error   func(args ...interface{})
	Errorln func(args ...interface{})
	Errorf  func(format string, args ...interface{})

	Fatal   func(args ...interface{})
	Fatalln func(args ...interface{})
	Fatalf  func(format string, args ...interface{})

	Panic   func(args ...interface{})
	Panicln func(args ...interface{})
	Panicf  func(format string, args ...interface{})

	Print   func(args ...interface{})
	Println func(args ...interface{})
	Printf  func(format string, args ...interface{})
)

/*
Initializes the Logrus logger
*/
func init() {
	// Set up logger
	log = logrus.New()

	// Set up shortcuts
	Trace = log.Trace
	Traceln = log.Traceln
	Tracef = log.Tracef

	Debug = log.Debug
	Debugln = log.Debugln
	Debugf = log.Debugf

	Info = log.Info
	Infoln = log.Infoln
	Infof = log.Infof

	Warn = log.Warn
	Warnln = log.Warnln
	Warnf = log.Warnf

	Error = log.Error
	Errorln = log.Errorln
	Errorf = log.Errorf

	Fatal = log.Fatal
	Fatalln = log.Fatalln
	Fatalf = log.Fatalf

	Panic = log.Panic
	Panicln = log.Panicln
	Panicf = log.Panicf

	Print = log.Print
	Println = log.Println
	Printf = log.Printf
}

/*
Log is the global Logrus logger instance
*/
func Log() *logrus.Logger {
	return log
}

/*
DefaultConfig initializes the Logrus logger in either debug or production mode.
*/
func DefaultConfig(l *logrus.Logger, debug bool) {
	if l == nil {
		l = Log()
	}
	if debug {
		l.SetLevel(logrus.TraceLevel)
		l.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})
	} else {
		l.SetLevel(logrus.WarnLevel)
		l.SetFormatter(&logrus.JSONFormatter{})
	}
}

/*
NoColors disables color formatting.
This only applies to the debug formatter. The Production formatter does not contain color by default.
*/
func NoColors(l *logrus.Logger) {
	if l == nil {
		l = Log()
	}
	if f, ok := l.Formatter.(*logrus.TextFormatter); ok {
		f.DisableColors = true
	}
}

/*
NoTimestamp removes the timestamp from the log.
*/
func NoTimestamp(l *logrus.Logger) {
	if l == nil {
		l = Log()
	}
	if f, ok := l.Formatter.(*logrus.TextFormatter); ok {
		f.DisableTimestamp = true
	} else if f, ok := l.Formatter.(*logrus.JSONFormatter); ok {
		f.DisableTimestamp = true
	}
}

/*
FullTimestamp ensures that the debug logger includes the full timestamp instead of a partial one.
The production logger outputs the full timestamp by default.
*/
func FullTimestamp(l *logrus.Logger) {
	if l == nil {
		l = Log()
	}
	if f, ok := l.Formatter.(*logrus.TextFormatter); ok {
		f.FullTimestamp = true
	}
}
