//go:build ignore

package app

import (
	"context"
	"oci-runtime/internal/app/mw"
	"oci-runtime/internal/infrastructure/technical/logging"
)

type HelloCmd struct{}

type Ports struct {
	Mount      MountManager
	// [...]
}

func NewHelloHandler(p Ports) mw.HandlerFunc[HelloCmd] {
	h := helloHandler{p}
	return h.handle
}

type helloHandler struct {
	p Ports
}

func (h *helloHandler) handle(ctx context.Context, _ InitCmd) error {
	l := logging.FromContext(ctx)

	l.Debug("Hello handler")

	return nil
}
