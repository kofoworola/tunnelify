package logging

import (
	"time"

	"github.com/kofoworola/tunnelify/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	*zap.Logger
}

func NewLogger(cfg *config.Config) (*Logger, error) {
	logPaths := append([]string{"stderr"}, cfg.Logging...)

	prodEncoderConfig := zap.NewProductionEncoderConfig()
	prodEncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339)

	config := zap.Config{
		Level:         zap.NewAtomicLevelAt(zapcore.WarnLevel),
		Encoding:      "json",
		EncoderConfig: prodEncoderConfig,
		OutputPaths:   logPaths,
	}
	if cfg.Debug {
		config.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	}

	logger, err := config.Build()
	if err != nil {
		return nil, err
	}
	return &Logger{logger}, nil
}

func (l *Logger) LogError(msg string, err error) {
	if err != nil {
		l.Logger.Error(
			msg,
			zapcore.Field{
				Key:    "error",
				String: err.Error(),
				Type:   zapcore.StringType,
			})
	}
}

func (l *Logger) With(key, val string) *Logger {
	return &Logger{
		l.Logger.With(zapcore.Field{
			Key:    key,
			String: val,
			Type:   zapcore.StringType,
		}),
	}
}

func (l *Logger) Warn(msg string, err error) {
	if err == nil {
		l.Logger.Warn(msg)
		return
	}
	l.Logger.Warn(
		msg,
		zapcore.Field{
			Key:    "error",
			String: err.Error(),
			Type:   zapcore.StringType,
		})

}
