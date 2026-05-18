package container

import (
	"context"
	"errors"
	"io/fs"
	"oci-runtime/internal/domain"
	"oci-runtime/internal/infrastructure/technical/xerr"
	"os"
	"path/filepath"
)

type Manager struct{}

func NewManager() *Manager {
	return &Manager{}
}

func (m *Manager) Load(ctx context.Context, root, name string) (domain.ContainerState, error) {
	fifoPath := filepath.Join(root, name, "exec.fifo")
	if _, err := os.Stat(fifoPath); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return domain.ContainerState{}, xerr.Op("container not found", err, xerr.KV{
				"root":      root,
				"name":      name,
				"exec_fifo": fifoPath,
			})
		}
		return domain.ContainerState{}, xerr.Op("stat exec fifo", err, xerr.KV{
			"exec_fifo": fifoPath,
		})
	}
	return domain.ContainerState{Name: name}, nil
}

func (m *Manager) Remove(ctx context.Context, root, name string) error {
	stateDir := filepath.Join(root, name)
	if err := os.RemoveAll(stateDir); err != nil {
		return xerr.Op("remove state dir", err, xerr.KV{
			"path": stateDir,
		})
	}
	return nil
}
