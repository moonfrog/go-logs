package logs

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime/debug"
	"time"

	"github.com/moonfrog/go-metrics"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	DEBUG = iota
	INFO
	WARN
	ERROR
	PANIC
	FATAL
)

const (
	DefaultBaseDir = "/var/moonfrog/go/"
)

type Level int

var currentLevel Level = INFO
var consoleLoggingEnabled bool = false
var consoleLogger *log.Logger = log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile)

var logger *log.Logger = nil

// sets the log file name to appName.log in defaultBaseDir
func InitDefault(appName string) {
	Init(appName, DefaultBaseDir)
}

// sets the log file name to appName.log in base dir
// takes the current time if appName is empty
func Init(appName string, baseDir string) {
	if baseDir == "" {
		baseDir = DefaultBaseDir
	}
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		log.Printf("E! could not create log directory. (%v)", err.Error())
		return
	}

	var logFileName string
	if appName == "" {
		logFileName = fmt.Sprintf("%v.log", time.Now().UnixNano())
	} else {
		logFileName = fmt.Sprintf("%v.log", appName)
	}

	logFilePath := filepath.Join(baseDir, logFileName)
	initLogger(logFilePath, 500, 10, 28)
}

func initLogger(logFilePath string, maxSize, maxBackups, maxAge int) {
	logger = log.New(&lumberjack.Logger{
		Filename:   logFilePath,
		MaxSize:    maxSize, // megabytes
		MaxBackups: maxBackups,
		MaxAge:     maxAge, //days
	}, "", log.LstdFlags|log.Lshortfile)
}

func SetConsoleLogging(status bool) {
	consoleLoggingEnabled = status
}

func SetLevel(level string) {
	switch level {
	case "info":
		currentLevel = INFO
	case "warn":
		currentLevel = WARN
	case "debug":
		currentLevel = DEBUG
	case "error":
		currentLevel = ERROR
	default:
		currentLevel = INFO
		Infof("wrong level %s: default info", level)
	}
}

func GetLevel() string {
	switch currentLevel {
	case DEBUG:
		return "debug"
	case WARN:
		return "warn"
	case INFO:
		return "info"
	case ERROR:
		return "error"
	}
	return "wrong"
}

func Log(level Level, v ...interface{}) {
	if level >= currentLevel {
		str := ""
		switch level {
		case DEBUG:
			str = "DEBUG"
		case INFO:
			str = "INFO"
		case WARN:
			str = "WARN"
		case ERROR:
			str = "ERROR"
		case PANIC:
			str = "PANIC"
		case FATAL:
			str = "FATAL"
		default:
			str = "INFO"
		}
		first, isString := v[0].(string)
		remaining := v[1:]
		var output string
		if isString {
			output = fmt.Sprintf("["+str+"] "+first, remaining...)
		} else {
			slice := []interface{}{"[" + str + "]", first}
			slice = append(slice, remaining...)
			output = fmt.Sprintln(slice...)
		}
		if logger != nil {
			logger.Output(3, output)
		}
		if consoleLoggingEnabled {
			consoleLogger.Output(3, output)
		}
	}
}

func Debugf(v ...interface{}) {
	Log(DEBUG, v...)
}

func Infof(v ...interface{}) {
	Log(INFO, v...)
}

func Warnf(v ...interface{}) {
	Log(WARN, v...)
	metrics.Update(metrics.RSTAT_WARN, 1)
}

// silent warn
func SWarnf(v ...interface{}) { // doesn't update metric
	Log(WARN, v...)
}

func Errorf(v ...interface{}) {
	Log(ERROR, v...)
	metrics.Update(metrics.RSTAT_ERROR, 1)
}

// silent error
func SErrorf(v ...interface{}) { // doesn't update metric
	Log(ERROR, v...)
}

func Panicf(v ...interface{}) {
	Log(PANIC, v...)
	s := fmt.Sprint(v...)
	panic(s)
	metrics.Update(metrics.RSTAT_PANIC, 1)
}

// silent panic
func SPanicf(v ...interface{}) { // doesn't update metric
	Log(PANIC, v...)
	s := fmt.Sprint(v...)
	panic(s)
}

func Fatalf(v ...interface{}) {
	Log(FATAL, v...)
	debug.PrintStack()
	os.Exit(1)
}
