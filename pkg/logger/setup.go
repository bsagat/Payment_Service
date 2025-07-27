package logger

import (
	"context"
	"log/slog"
	"os"
)

const (
	prodLevel  = "prod"
	devLevel   = "dev"
	debugLevel = "debug"
)

type Logger interface {
	Info(ctx context.Context, msg string, args ...any)
	Error(ctx context.Context, msg string, args ...any)
	Warn(ctx context.Context, msg string, args ...any)
	Debug(ctx context.Context, msg string, args ...any)
	Fatal(ctx context.Context, msg string, args ...any)
	With(args ...any) Logger
}

type SLogger struct {
	log *slog.Logger
}

func New(level string) Logger {
	var h slog.Handler
	switch level {
	case prodLevel:
		h = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
	case devLevel:
		h = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})
	case debugLevel:
		h = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})
	default:
		h = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
	}

	return SLogger{
		log: slog.New(h),
	}
}

func (s SLogger) Info(ctx context.Context, msg string, args ...any) {
	s.log.InfoContext(ctx, msg, args...)
}

func (s SLogger) Error(ctx context.Context, msg string, args ...any) {
	s.log.ErrorContext(ctx, msg, args...)
}

func (s SLogger) Warn(ctx context.Context, msg string, args ...any) {
	s.log.WarnContext(ctx, msg, args...)
}

func (s SLogger) Debug(ctx context.Context, msg string, args ...any) {
	s.log.DebugContext(ctx, msg, args...)
}

func (s SLogger) Fatal(ctx context.Context, msg string, args ...any) {
	s.log.ErrorContext(ctx, msg, args...)
	os.Exit(1)
}

func (s SLogger) With(args ...any) Logger {
	return SLogger{
		log: s.log.With(args...),
	}
}
