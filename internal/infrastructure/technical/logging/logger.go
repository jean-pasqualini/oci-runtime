package logging

import (
	"fmt"
	"github.com/lmittmann/tint"
	//"github.com/samber/slog-multi"
	"log/slog"
	"os"
	"time"
)

type Logger = slog.Logger

func New(name string, path string) (*Logger, error) {

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
	//slogmulti.Fanout(consoleHandler, fileHandler)
	return slog.New(consoleHandler), nil
}
