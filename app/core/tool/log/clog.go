package clog

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"time"
)

var eventLog *logrus.Logger
var verboseLog *logrus.Logger

var logBaseDir = "/opt/data/cc-robot/runtime/logs"

func EventLog() *logrus.Logger {
	if eventLog != nil {
		return eventLog
	}
	eventLog = logrus.New()
	file, err := os.OpenFile(fmt.Sprintf("%s/app-event.log", logBaseDir), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		eventLog.SetOutput(file)
	}
	return eventLog
}

func VerboseLog() *logrus.Logger {
	if verboseLog != nil {
		return verboseLog
	}
	verboseLog = logrus.New()
	file, err := os.OpenFile(fmt.Sprintf("%s/app-verbose.log", logBaseDir), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		verboseLog.SetOutput(file)
	}
	return verboseLog
}

func InitLogOptions(log *logrus.Logger, isDev bool) {
	log.SetReportCaller(true)
	formatter := &logrus.TextFormatter{
		FullTimestamp: true,
		TimestampFormat: time.RFC3339Nano,
	}
	log.SetFormatter(formatter)

	logLevel := logrus.InfoLevel
	if isDev {
		logLevel = logrus.DebugLevel
	}
	log.SetLevel(logLevel)
}