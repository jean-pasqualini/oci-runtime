package app

import (
	"context"
	"oci-runtime/internal/app/mw"
)

type CreateCmd struct {
	Name          string
	StatePath     string
	BundlePath    string
	LogPath       string
	LogFormat     string
	PidFile       string
	ConsoleSocket string
}

func NewCreateHandler() mw.HandlerFunc[CreateCmd] {
	h := createHandler{}
	return h.handle
}

type createHandler struct {
	p Ports
}

func (h *createHandler) handle(ctx context.Context, cmd CreateCmd) error {
	return nil
}
