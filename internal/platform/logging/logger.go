package logging

import (
	"fmt"
	"github.com/lmittmann/tint"
	"log/slog"
	"os"
	"time"
)

type Logger = slog.Logger

func New(name string) *Logger {
	h := tint.NewHandler(os.Stdout, &tint.Options{
		TimeFormat: time.RFC3339,
		Level:      slog.LevelInfo,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.MessageKey {
				a.Value = slog.StringValue(fmt.Sprintf("[%s] %s", name, a.Value.String()))
			}
			return a
		},
	})
	return slog.New(h)
}
