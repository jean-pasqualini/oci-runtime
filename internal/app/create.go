package app

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/creack/pty"
	"golang.org/x/sys/unix"
	"io"
	"oci-runtime/internal/app/mw"
	"oci-runtime/internal/domain"
	"oci-runtime/internal/infrastructure/technical/logging"
	"oci-runtime/internal/infrastructure/technical/xerr"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

type CreateCmd struct {
	Name          string
	MetadataRoot  string
	BundleRoot    string
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

func (h *createHandler) withTTY(attr syscall.SysProcAttr, fd int) syscall.SysProcAttr {
	// setsid + TIOCSCTTY needs to be done by the child
	attr.Setsid = true
	attr.Setctty = true
	attr.Ctty = 0

	return attr
}

func (h *createHandler) sendFDOverSocket(socket string, fd int) error {
	conn, err := unix.Socket(unix.AF_UNIX, unix.SOCK_STREAM, 0)
	if err != nil {
		return err
	}
	defer unix.Close(conn)

	addr := &unix.SockaddrUnix{Name: socket}
	if err := unix.Connect(conn, addr); err != nil {
		return err
	}

	// out of band data
	oob := unix.UnixRights(fd)
	if err := unix.Sendmsg(conn, []byte("x"), oob, nil, 0); err != nil {
		return err
	}
	return nil
}

func (h *createHandler) createInitCmdline() []string {
	args := os.Args[1:]

	// 2) by default keep all
	cut := args

	// 3) cut after "create"
	for i, a := range args {
		if a == "create" || a == "run" {
			cut = args[:i]
			break
		}
	}

	// 4) append "run"
	cut = append(cut, "init")

	return cut
}

func (h *createHandler) creatingInit(ctx context.Context, pidFile string, consoleSocket string, containerMedadataRoot string) (Ipc, error) {
	l := logging.FromContext(ctx)

	// Bidirectional SYNC_PIPE
	initRead, ociWrite, err := os.Pipe()
	if err != nil {
		return nil, xerr.Op("bidirectional SYNC_PIPE", err, xerr.KV{})
	}
	ociRead, initWrite, err := os.Pipe()
	if err != nil {
		return nil, xerr.Op("bidirectional SYNC_PIPE", err, xerr.KV{})
	}

	// Unidirectional KICK OFF pipe
	execFifoPath := filepath.Join(containerMedadataRoot, "exec.fifo")
	if err := unix.Mkfifo(execFifoPath, 0622); err != nil {
		return nil, xerr.Op("create exec fifo", err, xerr.KV{
			"exec_fifo_path": execFifoPath,
		})
	}
	execFifoFD, err := unix.Open(execFifoPath, unix.O_PATH|unix.O_CLOEXEC, 0)
	if err != nil {
		return nil, xerr.Op("unable to open in path mode the exec fifo", err, xerr.KV{
			"exec_fifo_path": execFifoPath,
		})
	}
	containerCommand := exec.CommandContext(
		ctx,
		"/proc/self/exe",
		h.createInitCmdline()...,
	)
	l.Error("cmdline init", "name", containerCommand.Path, "args", strings.Join(containerCommand.Args, " "))
	var consoleMaster *os.File = nil
	if consoleSocket != "" {
		master, consoleSlave, err := pty.Open()
		consoleMaster = master
		if err != nil {
			return nil, err
		}
		defer consoleMaster.Close()
		defer consoleSlave.Close()

		containerCommand.Stdout = consoleSlave
		containerCommand.Stdin = consoleSlave
		containerCommand.Stderr = consoleSlave
	} else {
		containerCommand.Stdout = os.Stdout
		containerCommand.Stdin = os.Stdin
		containerCommand.Stderr = os.Stderr
	}

	containerAttributes := syscall.SysProcAttr{}
	containerAttributes = h.withNamespace(containerAttributes)
	if consoleSocket != "" {
		containerAttributes = h.withTTY(containerAttributes, int(consoleMaster.Fd()))
	}
	containerCommand.SysProcAttr = &containerAttributes

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
	l.Debug("start fork to run init")
	if err := containerCommand.Start(); err != nil {
		return nil, xerr.Op("start fork to run init", err, xerr.KV{})
	}
	if pidFile != "" {
		if err := os.WriteFile(pidFile, []byte(fmt.Sprintf("%d", containerCommand.Process.Pid)), 0644); err != nil {
			return nil, xerr.Op("write process pid file", err, xerr.KV{})
		}
	}
	if err := initRead.Close(); err != nil {
		return nil, err
	}
	if err := initWrite.Close(); err != nil {
		return nil, err
	}
	if consoleMaster != nil {
		if err := h.sendFDOverSocket(consoleSocket, int(consoleMaster.Fd())); err != nil {
			return nil, err
		}
	}

	return h.ipcFactory(ociRead, ociWrite), err
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

	containerStateFolder := filepath.Join(cmd.MetadataRoot, cmd.Name)

	// Check if state already exists
	if _, err := os.Stat(containerStateFolder); err == nil {
		return xerr.Op("state folder already exists", fmt.Errorf("check"), xerr.KV{
			"state_directory": containerStateFolder,
		})
	}

	// Create state
	if err := os.MkdirAll(containerStateFolder, 0777); err != nil {
		return xerr.Op("make state path directory", err, xerr.KV{
			"state_directory": containerStateFolder,
		})
	}

	logger.Info("init logfile path", "path", cmd.LogPath)

	syncPipe, err := h.creatingInit(ctx, cmd.PidFile, cmd.ConsoleSocket, containerStateFolder)
	if err != nil {
		return xerr.Op("creating init", err, xerr.KV{})
	}
	defer syncPipe.Close()

	logger.Info("load config file")
	containerConfigFile, err := os.Open(filepath.Join(cmd.BundleRoot, "config.json"))
	if err != nil {
		return err
	}
	containerConfig, err := io.ReadAll(containerConfigFile)
	if err != nil {
		return err
	}
	logger.Info("send config file to init process")
	if err := syncPipe.Send(ctx, json.RawMessage(containerConfig)); err != nil {
		return err
	}

	logger.Info("waiting for init process bootstrap")
	var initDone bool
	if err := syncPipe.Recv(ctx, &initDone); err != nil {
		return err
	}
	logger.Info("init process bootstraped")

	return nil
}
