package logging

import (
	"fmt"
	"github.com/lmittmann/tint"
	"github.com/samber/slog-multi"
	"log/slog"
	"os"
	"time"
)

type Logger = slog.Logger

func New(name string, path string) (*Logger, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, err
	}

	fileHandler := slog.NewJSONHandler(f, &slog.HandlerOptions{
		Level: slog.LevelInfo,
		// AddSource: true, // optionnel: ajoute fichier:ligne
	})

	consoleHandler := tint.NewHandler(os.Stdout, &tint.Options{
		TimeFormat: time.RFC3339,
		Level:      slog.LevelDebug,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.MessageKey {
				a.Value = slog.StringValue(fmt.Sprintf("[%s] %s", name, a.Value.String()))
			}
			return a
		},
	})
	return slog.New(slogmulti.Fanout(consoleHandler, fileHandler)), nil
}
