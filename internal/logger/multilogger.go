package logger

import (
	"context"
	"errors"
	"sync"
)

type MultiLogger struct {
	loggers sync.Map
}

func NewMultiLogger() *MultiLogger {
	return &MultiLogger{loggers: sync.Map{}}
}

func (ml *MultiLogger) AddLoggerToMultiLogger(logger *Logger) {
	ml.loggers.Store(logger.Level, logger)
}

var ErrNoLoggerForThisLevel = errors.New("error, no logger for the request log level")
var ErrLoggerForThisLevelIsInvalid = errors.New("error, logger for the requested level is invalid")

func (ml *MultiLogger) Log(ctx context.Context, msg LogMessage) error {
	logger, ok := ml.loggers.Load(msg.Level)
	if !ok || logger == nil {
		return ErrNoLoggerForThisLevel
	}
	lg, ok := logger.(*Logger)
	if !ok {
		return ErrLoggerForThisLevelIsInvalid
	}
	return lg.Log(ctx, msg)
}
