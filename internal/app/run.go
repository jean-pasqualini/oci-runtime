package app

import (
	"context"
	"encoding/json"
	"io"
	"oci-runtime/internal/app/mw"
	"oci-runtime/internal/domain"
	"oci-runtime/internal/infrastructure/technical/logging"
	"oci-runtime/internal/infrastructure/transport/ipc"
	"oci-runtime/internal/infrastructure/transport/rpc"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

type RunCmd struct {
	Name          string
	StatePath     string
	BundlePath    string
	LogPath       string
	LogFormat     string
	PidFile       string
	ConsoleSocket string
}

func NewRunHandler() mw.HandlerFunc[RunCmd] {
	h := runHandler{}
	return h.handle
}

type runHandler struct {
}

func (h *runHandler) withNamespace(attr syscall.SysProcAttr) syscall.SysProcAttr {
	attr.Cloneflags = syscall.CLONE_NEWNS | // Mount namespace
		syscall.CLONE_NEWPID | // Pid namespace
		syscall.CLONE_NEWNET | // Network namespace
		syscall.CLONE_NEWTIME | // Time namespace
		syscall.CLONE_NEWUTS // Hostname namespace

	return attr
}

func (h *runHandler) startInit(ctx context.Context) (*exec.Cmd, Rpc, error) {
	initRead, ociWrite, err := os.Pipe()
	if err != nil {
		return nil, nil, err
	}
	ociRead, initWrite, err := os.Pipe()
	if err != nil {
		return nil, nil, err
	}

	containerCommand := exec.CommandContext(ctx, "/proc/self/exe", "init")
	containerAttributes := h.withNamespace(syscall.SysProcAttr{})
	containerCommand.SysProcAttr = &containerAttributes
	containerCommand.Stdout = os.Stdout
	containerCommand.Stdin = os.Stdin
	containerCommand.Stderr = os.Stderr
	containerCommand.Env = []string{
		"FD_SYNC_READ=3",
		"FD_SYNC_WRITE=4",
	}
	containerCommand.ExtraFiles = []*os.File{
		initRead,
		initWrite,
	}
	if err := containerCommand.Start(); err != nil {
		return nil, nil, err
	}
	if err := initRead.Close(); err != nil {
		return nil, nil, err
	}
	if err := initWrite.Close(); err != nil {
		return nil, nil, err
	}

	return nil, rpc.NewRpcPipe(ipc.NewSyncPipe(ociRead, ociWrite)), err
}

func (h *runHandler) fetchContainerConfig(ctx context.Context, fd *os.File) (domain.ContainerConfiguration, error) {
	containerConfig := domain.ContainerConfiguration{}
	dec := json.NewDecoder(fd)
	if err := dec.Decode(&containerConfig); err != nil {
		return domain.ContainerConfiguration{}, err
	}
	return containerConfig, nil
}

func (h *runHandler) handle(ctx context.Context, cmd RunCmd) error {
	logger := logging.FromContext(ctx)
	logger.Info("creating init", "RunCmd", cmd)

	_, syncPipe, err := h.startInit(ctx)
	if err != nil {
		return err
	}
	defer syncPipe.Close()

	containerConfigFile, err := os.Open(filepath.Join(cmd.BundlePath, "config.json"))
	if err != nil {
		return err
	}
	containerConfig, err := io.ReadAll(containerConfigFile)
	if err != nil {
		return err
	}

	syncPipe.Op("config", containerConfig)

	return nil // containerCommand.Wait()
}
