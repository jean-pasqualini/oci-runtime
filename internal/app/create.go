package app

import (
	"context"
	"encoding/json"
	"fmt"
	"golang.org/x/sys/unix"
	"io"
	"oci-runtime/internal/app/mw"
	"oci-runtime/internal/domain"
	"oci-runtime/internal/infrastructure/technical/logging"
	"oci-runtime/internal/infrastructure/technical/xerr"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
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

func NewCreateHandler(factory IpcFactory) mw.HandlerFunc[CreateCmd] {
	h := createHandler{ipcFactory: factory}
	return h.handle
}

type createHandler struct {
	ipcFactory IpcFactory
}

func (h *createHandler) withNamespace(attr syscall.SysProcAttr) syscall.SysProcAttr {
	attr.Cloneflags = syscall.CLONE_NEWNS | // Mount namespace
		syscall.CLONE_NEWPID | // Pid namespace
		syscall.CLONE_NEWNET | // Network namespace
		syscall.CLONE_NEWTIME | // Time namespace
		syscall.CLONE_NEWUTS // Hostname namespace

	return attr
}

func (h *createHandler) startInit(ctx context.Context, statePath string) (*exec.Cmd, Ipc, error) {
	l := logging.FromContext(ctx)
	// Bidirectional SYNC_PIPE
	initRead, ociWrite, err := os.Pipe()
	if err != nil {
		return nil, nil, err
	}
	ociRead, initWrite, err := os.Pipe()
	if err != nil {
		return nil, nil, err
	}

	// Unidirectional KICK OFF pipe
	execFifoPath := filepath.Join(statePath, "exec.fifo")
	if err := unix.Mkfifo(execFifoPath, 0622); err != nil {
		return nil, nil, xerr.Op("create exec fifo", err, xerr.KV{
			"exec_fifo_path": execFifoPath,
		})
	}
	execFifoFD, err := unix.Open(execFifoPath, unix.O_PATH|unix.O_CLOEXEC, 0)
	if err != nil {
		return nil, nil, xerr.Op("unable to open in path mode the exec fifo", err, xerr.KV{
			"exec_fifo_path": execFifoPath,
		})
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
		"FD_EXEC=5",
	}
	containerCommand.ExtraFiles = []*os.File{
		initRead,
		initWrite,
		os.NewFile(uintptr(execFifoFD), "exec.fifo"),
	}
	l.Info("Start containerCommand")
	if err := containerCommand.Start(); err != nil {
		return nil, nil, err
	}
	if err := initRead.Close(); err != nil {
		return nil, nil, err
	}
	if err := initWrite.Close(); err != nil {
		return nil, nil, err
	}
	return nil, h.ipcFactory(ociRead, ociWrite), err
}

func (h *createHandler) fetchContainerConfig(ctx context.Context, fd *os.File) (domain.ContainerConfiguration, error) {
	containerConfig := domain.ContainerConfiguration{}
	dec := json.NewDecoder(fd)
	if err := dec.Decode(&containerConfig); err != nil {
		return domain.ContainerConfiguration{}, err
	}
	return containerConfig, nil
}

func (h *createHandler) handle(ctx context.Context, cmd CreateCmd) error {
	logger := logging.FromContext(ctx)
	logger.Info("creating init", "RunCmd", cmd)

	// Check if state already exists
	if _, err := os.Stat(cmd.StatePath); err == nil {
		return xerr.Op("state folder already exists", fmt.Errorf("check"), xerr.KV{
			"state_directory": cmd.StatePath,
		})
	}

	// Create state
	if err := os.MkdirAll(cmd.StatePath, 0777); err != nil {
		return xerr.Op("make state path directory", err, xerr.KV{
			"state_directory": cmd.StatePath,
		})
	}

	_, syncPipe, err := h.startInit(ctx, cmd.StatePath)
	if err != nil {
		return xerr.Op("start init", err, xerr.KV{})
	}
	defer syncPipe.Close()

	logger.Info("load config file")
	containerConfigFile, err := os.Open(filepath.Join(cmd.BundlePath, "config.json"))
	if err != nil {
		return err
	}
	containerConfig, err := io.ReadAll(containerConfigFile)
	if err != nil {
		return err
	}
	logger.Info("send config file to init process")
	if err := syncPipe.Send(json.RawMessage(containerConfig)); err != nil {
		return err
	}

	logger.Info("waiting for init process bootstrap")
	var initDone bool
	if err := syncPipe.Recv(&initDone); err != nil {
		return err
	}
	logger.Info("init process bootstraped")

	logger.Info("oci runtime finished")
	return nil // containerCommand.Wait()
}
