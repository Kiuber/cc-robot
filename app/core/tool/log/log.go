package clog

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"log"
)

const eventLogKey = iota
const verboseLogKey = iota
var EventLog *zap.Logger
var VerboseLog *zap.Logger

var logBaseDir = "/opt/data/cc-robot/runtime/logs"

func InitEventLog(isDev bool) *zap.Logger {
	if EventLog != nil {
		return EventLog
	}

	logLevel := zap.InfoLevel
	if isDev {
		logLevel = zap.DebugLevel
	}

	core := zapcore.NewCore(
		DefaultZapJSONEncoder(),
		zapcore.AddSync(DefaultSyncWriter("app-event.log")),
		logLevel,
	)
	EventLog = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.WarnLevel))
	return EventLog
}

func InitVerboseLog(isDev bool) *zap.Logger {
	if VerboseLog != nil {
		return VerboseLog
	}

	logLevel := zap.InfoLevel
	if isDev {
		logLevel = zap.DebugLevel
	}

	core := zapcore.NewCore(
		DefaultZapJSONEncoder(),
		zapcore.AddSync(DefaultSyncWriter("app-verbose.log")),
		logLevel,
	)
	VerboseLog = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.WarnLevel))
	return VerboseLog
}

func DefaultSyncWriter(fileName string) zapcore.WriteSyncer {
	return zapcore.AddSync(&lumberjack.Logger{
		Filename:   buildLogPath(fileName),
		MaxSize:    32,
		MaxBackups: 1024,
		MaxAge:     365,
		Compress:   false,
		LocalTime:  true,
	})
}

func DefaultZapJSONEncoder() zapcore.Encoder {
	return zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	})
}

func buildLogPath(fileName string) string {
	path := fmt.Sprintf("%s/%s", logBaseDir, fileName)
	log.Printf("log path: %s", path)
	return path
}

func NewContext(ctx context.Context, fields ...zapcore.Field) context.Context {
	ctx = context.WithValue(ctx, eventLogKey, WithCtxEventLog(ctx).With(fields...))
	return context.WithValue(ctx, verboseLogKey, WithCtxVerboseLog(ctx).With(fields...))
}

func WithCtxEventLog(ctx context.Context) *zap.Logger {
	if ctx == nil {
		return EventLog
	}
	if ctxLogger, ok := ctx.Value(eventLogKey).(*zap.Logger); ok {
		return ctxLogger
	}
	return EventLog
}

func WithCtxVerboseLog(ctx context.Context) *zap.Logger {
	if ctx == nil {
		return VerboseLog
	}
	if ctxLogger, ok := ctx.Value(verboseLogKey).(*zap.Logger); ok {
		return ctxLogger
	}
	return VerboseLog
}