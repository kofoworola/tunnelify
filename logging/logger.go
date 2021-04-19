package logging

import (
	"time"

	"github.com/kofoworola/tunnelify/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	logger *zap.Logger
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

	logger, err := config.Build()
	if err != nil {
		return nil, err
	}
	return &Logger{logger}, nil
}

func (l *Logger) LogError(msg string, err error) {
	if err != nil {
		l.logger.Error(
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
		l.logger.With(zapcore.Field{
			Key:    key,
			String: val,
			Type:   zapcore.StringType,
		})}
}

func (l *Logger) Warn(msg string) {
	l.logger.Warn(msg)
}

// TODO combine this and above
func (l *Logger) WarnError(msg string, err error) {
	l.logger.Warn(
		msg,
		zapcore.Field{
			Key:    "error",
			String: err.Error(),
			Type:   zapcore.StringType,
		})
}
