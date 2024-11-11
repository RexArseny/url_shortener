package logger

import (
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func InitLogger() (*zap.Logger, error) {
	config := zap.NewProductionConfig()

	config.EncoderConfig.TimeKey = "time"
	config.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339)

	return config.Build()
}
