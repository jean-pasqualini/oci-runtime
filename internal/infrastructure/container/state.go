package container

import (
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"oci-runtime/internal/domain"
	"oci-runtime/internal/infrastructure/technical/xerr"
	"os"
	"path/filepath"
	"strconv"
)

type Manager struct{}

func NewManager() *Manager {
	return &Manager{}
}

func (m *Manager) Load(ctx context.Context, root, name string) (domain.ContainerState, error) {
	statePath := filepath.Join(root, name, "state.json")
	data, err := os.ReadFile(statePath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return domain.ContainerState{}, xerr.Op("container not found", err, xerr.KV{
				"root":       root,
				"name":       name,
				"state_file": statePath,
			})
		}
		return domain.ContainerState{}, xerr.Op("read state", err, xerr.KV{
			"state_file": statePath,
		})
	}

	var state domain.ContainerState
	if err := json.Unmarshal(data, &state); err != nil {
		return domain.ContainerState{}, xerr.Op("unmarshal state", err, xerr.KV{
			"state_file": statePath,
		})
	}

	fifoPath := filepath.Join(root, name, "exec.fifo")
	if _, err := os.Stat(fifoPath); err == nil {
		state.Status = domain.StatusCreated
		return state, nil
	} else if !errors.Is(err, fs.ErrNotExist) {
		return domain.ContainerState{}, xerr.Op("stat exec fifo", err, xerr.KV{
			"exec_fifo": fifoPath,
		})
	}

	procPath := filepath.Join("/proc", strconv.Itoa(state.Pid))
	if _, err := os.Stat(procPath); err == nil {
		state.Status = domain.StatusRunning
	} else if errors.Is(err, fs.ErrNotExist) {
		state.Status = domain.StatusStopped
	} else {
		return domain.ContainerState{}, xerr.Op("stat proc pid", err, xerr.KV{
			"proc_path": procPath,
			"pid":       strconv.Itoa(state.Pid),
		})
	}

	return state, nil
}

func (m *Manager) Save(ctx context.Context, root string, state domain.ContainerState) error {
	statePath := filepath.Join(root, state.Name, "state.json")

	data, err := json.Marshal(state)
	if err != nil {
		return xerr.Op("marshal state", err, xerr.KV{
			"path": statePath,
		})
	}
	if err := os.WriteFile(statePath, data, 0o644); err != nil {
		return xerr.Op("write state", err, xerr.KV{
			"path": statePath,
		})
	}
	return nil
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
