package clog

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var eventLog *zap.Logger
var verboseLog *zap.Logger

var logBaseDir = "/opt/data/cc-robot/runtime/logs"

func EventLog() *zap.Logger {
	if eventLog != nil {
		return eventLog
	}

	core := zapcore.NewCore(
		DefaultZapJSONEncoder(),
		zapcore.AddSync(DefaultSyncWriter("app-event.log")),
		zap.DebugLevel,
	)
	eventLog = zap.New(core, zap.AddCaller())
	return eventLog
}

func VerboseLog() *zap.Logger {
	if verboseLog != nil {
		return verboseLog
	}

	core := zapcore.NewCore(
		DefaultZapJSONEncoder(),
		zapcore.AddSync(DefaultSyncWriter("app-verbose.log")),
		zap.DebugLevel,
	)
	verboseLog = zap.New(core, zap.AddCaller())
	return verboseLog
}

func DefaultSyncWriter(fileName string) zapcore.WriteSyncer {
	return zapcore.AddSync(&lumberjack.Logger{
		Filename:   fmt.Sprintf("%s/%s", logBaseDir, fileName),
		MaxSize:    1024,
		MaxBackups: 200,
		MaxAge:     365,
		Compress:   false,
		LocalTime:  true,
	})
}

func DefaultZapJSONEncoder() zapcore.Encoder {
	return zapcore.NewJSONEncoder(zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	})
}
