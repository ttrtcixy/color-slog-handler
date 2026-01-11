package logger

import (
	"log/slog"
)

func levelColor(l slog.Level) []byte {
	switch l {
	case slog.LevelDebug:
		return blue
	case slog.LevelInfo:
		return green
	case slog.LevelWarn:
		return yellow
	case slog.LevelError:
		return red
	default:
		return none
	}
}

func ParseLevel(level int) string {
	switch slog.Level(level) {
	case slog.LevelDebug:
		return "DEBUG"
	case slog.LevelInfo:
		return "INFO"
	case slog.LevelWarn:
		return "WARN"
	case slog.LevelError:
		return "ERROR"
	default:
		return "INFO"
	}
}

func levelBytes(level slog.Level) string {
	switch level {
	case slog.LevelDebug:
		return LevelDebug
	case slog.LevelInfo:
		return LevelInfo
	case slog.LevelWarn:
		return LevelWarn
	case slog.LevelError:
		return LevelError
	default:
		return LevelInfo
	}
}
