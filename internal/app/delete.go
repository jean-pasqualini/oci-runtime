package app

import (
	"context"
	"oci-runtime/internal/app/mw"
	"oci-runtime/internal/infrastructure/technical/logging"
	"oci-runtime/internal/infrastructure/technical/xerr"
)

type DeleteCmd struct {
	Name         string
	MetadataRoot string
}

func NewDeleteHandler(state ContainerStateLoader) mw.HandlerFunc[DeleteCmd] {
	h := deleteHandler{state: state}
	return h.handle
}

type deleteHandler struct {
	state ContainerStateLoader
}

func (h *deleteHandler) handle(ctx context.Context, c DeleteCmd) error {
	l := logging.FromContext(ctx)
	l.With("name", c.Name).Debug("Delete handler")

	_, err := h.state.Load(ctx, c.MetadataRoot, c.Name)
	if err != nil {
		return xerr.Op("load container state", err, xerr.KV{
			"name": c.Name,
			"root": c.MetadataRoot,
		})
	}

	// TODO(task 3): refuse if running
	// TODO(task 4): linux teardown (mounts, namespaces, cgroup/proc)
	// TODO(task 5): ipc transport teardown
	// TODO(task 6): remove --root/<name> state directory last

	return nil
}
