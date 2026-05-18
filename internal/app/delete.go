package app

import (
	"context"
	"errors"
	"oci-runtime/internal/app/mw"
	"oci-runtime/internal/domain"
	"oci-runtime/internal/infrastructure/technical/logging"
	"oci-runtime/internal/infrastructure/technical/xerr"
)

type DeleteCmd struct {
	Name         string
	MetadataRoot string
}

func NewDeleteHandler(state ContainerStateManager) mw.HandlerFunc[DeleteCmd] {
	h := deleteHandler{state: state}
	return h.handle
}

type deleteHandler struct {
	state ContainerStateManager
}

func (h *deleteHandler) handle(ctx context.Context, c DeleteCmd) error {
	l := logging.FromContext(ctx)
	l.With("name", c.Name).Debug("Delete handler")

	state, err := h.state.Load(ctx, c.MetadataRoot, c.Name)
	if err != nil {
		return xerr.Op("load container state", err, xerr.KV{
			"name": c.Name,
			"root": c.MetadataRoot,
		})
	}

	if state.Status == domain.StatusRunning {
		return xerr.Op("refuse delete", errors.New("container is running"), xerr.KV{
			"name":   c.Name,
			"status": state.Status,
		})
	}

	if err := h.state.Remove(ctx, c.MetadataRoot, c.Name); err != nil {
		return err
	}

	l.With("name", c.Name).Info("container deleted")
	return nil
}
