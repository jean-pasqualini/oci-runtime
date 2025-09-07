package mw

import (
	"context"
)

type HandlerFunc[C any] func(ctx context.Context, cmd C) error
type Middleware[C any] func(HandlerFunc[C]) HandlerFunc[C]

func Chain[C any](h HandlerFunc[C], mws ...Middleware[C]) HandlerFunc[C] {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}
