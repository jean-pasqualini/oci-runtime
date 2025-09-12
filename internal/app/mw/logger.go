package mw

import (
	"context"
	"fmt"
	"log/slog"
	"oci-runtime/internal/infrastructure/technical/logging"
	"oci-runtime/internal/infrastructure/technical/xerr"
	"os"
)

func WithLogging[C any](name string, logger *slog.Logger) Middleware[C] {
	return func(next HandlerFunc[C]) HandlerFunc[C] {
		return func(ctx context.Context, cmd C) error {
			//start := time.Now()
			logger = logger.With(
				"command", fmt.Sprintf("%T", cmd),
				"pid", os.Getpid(),
			)
			logger.Info(
				"info about the process",
				"ppid", os.Getppid(),
				"args", os.Args,
			)
			ctx = logging.WithLogger(
				ctx,
				logger,
			)
			if err := next(ctx, cmd); err != nil {
				var attrs []any
				if xe, ok := err.(xerr.AttrLog); ok {
					attrs = xe.LogAttrs()
				}
				logger.With(attrs...).Error(err.Error())
				return err
			}
			return nil
		}
	}
}
