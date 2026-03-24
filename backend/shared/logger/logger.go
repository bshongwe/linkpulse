package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.Logger

func Init(env string) {
	var config zap.Config

	if env == "production" || env == "prod" {
		config = zap.NewProductionConfig()
	} else {
		config = zap.NewDevelopmentConfig()
	}

	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	var err error
	Log, err = config.Build(zap.AddCallerSkip(1), zap.WithCaller(true))
	if err != nil {
		panic("failed to initialize logger: " + err.Error())
	}
}

func WithTraceID(traceID string) *zap.Logger {
	if Log == nil {
		return zap.NewNop()
	}
	return Log.With(zap.String("trace_id", traceID))
}
