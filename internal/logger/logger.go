package logger

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/JuanX-G/crossbow"
)

type ctxKey string

const LoggerKey ctxKey = "logger"

func InjectLoggerMiddleware(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), LoggerKey, logger)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

type Logger struct {
	Level LogLevel
	srv   *crossbow.Server[*LogHandler, LogMessage, any]
}

func InitLogger(level LogLevel, dest io.WriteCloser) (*Logger, error) {
	cfg := crossbow.ServerConfig{}
	if level == LogLevelInfo || level == LogLevelDebug {
		cfg.Policy = crossbow.PolicyDropNewest
		cfg.Workers = 2
	} else {
		cfg.Policy = crossbow.PolicyBlock
		cfg.Workers = 4
	}
	cfg.MailboxSize = 64

	handler := &LogHandler{Dest: dest}

	srv, err := crossbow.NewServer(handler, cfg, crossbow.DefaultPanicRecover)
	if err != nil {
		return nil, err
	}

	return &Logger{
		Level: level,
		srv:   srv,
	}, nil
}

var ErrMismatchedLevel = errors.New("error mismatched level")

func (l *Logger) Log(ctx context.Context, msg LogMessage) error {
	if msg.Level != l.Level {
		return ErrMismatchedLevel
	}
	return l.srv.Send(ctx, msg)
}

func MakeDefaultMultiLogger(dest io.WriteCloser) (*MultiLogger, error) {
	mlogger := NewMultiLogger()

	LOGLEVELS := []LogLevel{LogLevelInfo, LogLevelDebug, LogLevelError}
	for _, level := range LOGLEVELS {
		lg, err := InitLogger(level, dest)
		if err != nil {
			return nil, err
		}
		mlogger.AddLoggerToMultiLogger(lg)
	}
	return mlogger, nil
}

type ScopedLogger struct {
	logger *MultiLogger
	source string
}

func NewScopedLogger(mlogger *MultiLogger, source string) *ScopedLogger {
	return &ScopedLogger{logger: mlogger, source: source}
}

func (sl *ScopedLogger) Log(ctx context.Context, msg string, level LogLevel) {
	sl.logger.Log(ctx, LogMessage{Level: level, Msg: msg, Source: sl.source})
}

func (sl *ScopedLogger) LogFrom(ctx context.Context, place, msg string, level LogLevel) {
	sl.logger.Log(ctx, LogMessage{Level: level, Msg: msg, Source: sl.source + "::" + place})
}
