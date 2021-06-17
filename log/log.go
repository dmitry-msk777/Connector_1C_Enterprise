package log

import (
	"errors"
	"fmt"
	"log"
	"sqlquerybuilder/config"
	"sqlquerybuilder/version"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/sirupsen/logrus"
)

var (
	level Level
)

const (
	// DebugLevel defines debug log level.
	DebugLevel Level = iota
	// InfoLevel defines info log level.
	InfoLevel
	// WarnLevel defines warn log level.
	WarnLevel
	// ErrorLevel defines error log level.
	ErrorLevel
	// FatalLevel defines fatal log level.
)

// Level defines log levels.
type Level int8

// SetLevel метод для установки логера
func SetLevel(l Level) {
	level = l
}

func print(pref, format string) {
	fmt.Println(pref, format)
}

func should(lvl Level) bool {
	if lvl > level || lvl == level {
		return true
	}
	return false
}

func getPrefix(l string) string {
	return fmt.Sprint(l, " ", time.Now().Format("2006/01/02 - 15:04:05"), " |")
}

type LogImpl interface {
	//InitLog()
	Info(args ...interface{})
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	Warningf(format string, args ...interface{})
	Debugf(format string, args ...interface{})
	Output(calldepth int, s string) error
	Panic() // Выполняется перед panic
}

var Impl LogImpl

func SetLogger(LogImpl LogImpl) {
	Impl = LogImpl
}

func InitLog(Config *config.Config) {

	switch Config.LogLevel {
	case 0:
		SetLevel(DebugLevel)
	case 1:
		SetLevel(InfoLevel)
	case 3:
		SetLevel(WarnLevel)
	case 4:
		SetLevel(ErrorLevel)
	default:
		SetLevel(DebugLevel)
	}

	switch Config.LoggerDefault {
	case "Sentry":
		Logger, err := NewSentryLog()
		if err != nil {
			Logger := NewStandartLog()
			SetLogger(Logger)
		} else {
			SetLogger(Logger)

		}
	case "Lorgus":
		Logger := NewLorgus()
		SetLogger(Logger)
	default:
		Logger := NewStandartLog()
		SetLogger(Logger)
	}

}

// Lorgus
type LorgusLog struct {
	Logger *logrus.Logger
}

func NewLorgus() *LorgusLog {
	LorgusLogStruct := &LorgusLog{
		Logger: logrus.New(),
	}

	LorgusLogStruct.Logger.Formatter = &logrus.JSONFormatter{}

	return LorgusLogStruct
}

func (LorgusLog LorgusLog) Error(args ...interface{}) {
	LorgusLog.Logger.Error(args...)
}

func (LorgusLog LorgusLog) Errorf(format string, args ...interface{}) {
	LorgusLog.Logger.Errorf(format, args...)
}

func (LorgusLog LorgusLog) Debugf(format string, args ...interface{}) {
	LorgusLog.Logger.Debugf(format, args...)
}

func (LorgusLog LorgusLog) Warningf(format string, args ...interface{}) {
	LorgusLog.Logger.Warnf(format, args...)
}

func (LorgusLog LorgusLog) Info(args ...interface{}) {
	LorgusLog.Logger.Info(args...)
}

func (LorgusLog LorgusLog) Output(calldepth int, s string) error {

	if calldepth == 4 {
		LorgusLog.Logger.Error(s)
	} else {
		LorgusLog.Logger.Info(s)
	}

	return nil
}

func (LorgusLog LorgusLog) Panic() {

}

// Lorgus END

// Standart Log
type StandartLog struct{}

func NewStandartLog() *StandartLog {
	StandartLogStruct := &StandartLog{}

	return StandartLogStruct
}

func (StandartLog StandartLog) Error(args ...interface{}) {
	if ok := should(DebugLevel); ok {
		s := fmt.Sprint(args...)
		print(getPrefix("[ERR]"), s)
	}
}

func (StandartLog StandartLog) Errorf(format string, args ...interface{}) {
	if ok := should(ErrorLevel); ok {
		s := fmt.Sprintf(format, args...)
		print(getPrefix("[ERR]"), s)
	}
}

func (StandartLog StandartLog) Info(args ...interface{}) {
	if ok := should(DebugLevel); ok {
		s := fmt.Sprint(args...)
		print(getPrefix("[INF]"), s)
	}
}

func (StandartLog StandartLog) Debugf(format string, args ...interface{}) {
	if ok := should(DebugLevel); ok {
		s := fmt.Sprintf(format, args...)
		print(getPrefix("[DEB]"), s)
	}
}

func (StandartLog StandartLog) Warningf(format string, args ...interface{}) {
	if ok := should(WarnLevel); ok {
		s := fmt.Sprintf(format, args...)
		print(getPrefix("[WAR]"), s)
	}
}

// TODO: При установке в NSQ logger всегда приходит уровень 2, понять почему
func (StandartLog StandartLog) Output(calldepth int, s string) error {

	if calldepth == 4 {
		log.Print(s)
	} else {
		log.Print(s)
	}

	return nil
}

func (StandartLog StandartLog) Panic() {

}

// Standart END

// Sentry log
type SentryLog struct{}

func NewSentryLog() (*SentryLog, error) {
	StandartLogStruct := &SentryLog{}

	err := sentry.Init(sentry.ClientOptions{
		Dsn:     config.Conf.SentryUrlDSN,
		Release: version.Commit + " - " + version.BuildTime,
		Debug:   true,
	})

	fmt.Println(config.Conf.SentryUrlDSN)

	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	return StandartLogStruct, nil
}

func (SentryLog SentryLog) Error(args ...interface{}) {

	StringError := fmt.Sprint(args...)
	// TODO: Создавать ошибку из текста как-то много лишнего.
	sentry.CaptureException(errors.New(StringError))

	print(getPrefix("[ERR]"), StringError)

}

func (SentryLog SentryLog) Info(args ...interface{}) {

	if ok := should(InfoLevel); ok {
		s := fmt.Sprint(args...)
		print(getPrefix("[INF]"), s)
	}
}

func (SentryLog SentryLog) Errorf(format string, args ...interface{}) {

	StringError := fmt.Sprintf(format, args...)
	// TODO: Создавать ошибку из текста как-то много лишнего.
	sentry.CaptureException(errors.New(StringError))
	print(getPrefix("[ERR]"), StringError)

}

func (SentryLog SentryLog) Debugf(format string, args ...interface{}) {
	if ok := should(DebugLevel); ok {
		s := fmt.Sprintf(format, args...)
		print(getPrefix("[INF]"), s)
	}
}

func (SentryLog SentryLog) Warningf(format string, args ...interface{}) {
	if ok := should(WarnLevel); ok {
		s := fmt.Sprintf(format, args...)
		print(getPrefix("[WAR]"), s)
	}
}

func (SentryLog SentryLog) Output(calldepth int, s string) error {

	if calldepth == 4 {
		sentry.CaptureException(errors.New(s))
		print(getPrefix("[ERR]"), s)
	} else {

		if ok := should(InfoLevel); ok {
			//s := fmt.Sprint(args...)
			print(getPrefix(""), s)
		}
	}

	return nil
}

func (SentryLog SentryLog) Panic() {
	sentry.Flush(time.Second * 5)
}
