package app

import (
	"context"
	"oci-runtime/internal/app/mw"
)

type StartCmd struct {
	Name      string
	StatePath string
}

func NewStartHandler() mw.HandlerFunc[StartCmd] {
	h := startHandler{}
	return h.handle
}

type startHandler struct {
	p Ports
}

func (h *startHandler) handle(ctx context.Context, cmd StartCmd) error {
	return nil
}
