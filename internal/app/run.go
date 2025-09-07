package app

import (
	"context"
	"oci-runtime/internal/app/mw"
	"oci-runtime/internal/platform/logging"
	"os"
	"os/exec"
	"syscall"
)

type RunCmd struct {
	Name string
}

func NewRunHandler() mw.HandlerFunc[RunCmd] {
	h := runHandler{}
	return h.handle
}

type runHandler struct {
}

func withNamespace(attr syscall.SysProcAttr) syscall.SysProcAttr {
	attr.Cloneflags = syscall.CLONE_NEWNS | // Mount namespace
		syscall.CLONE_NEWPID | // Pid namespace
		syscall.CLONE_NEWNET | // Network namespace
		syscall.CLONE_NEWTIME | // Time namespace
		syscall.CLONE_NEWUTS // Hostname namespace

	return attr
}

func (r *runHandler) handle(ctx context.Context, cmd RunCmd) error {
	logger := logging.FromContext(ctx)
	logger.Info("creating init", "RunCmd", cmd)

	containerCommand := exec.CommandContext(ctx, "/proc/self/exe", "init")
	containerAttributes := withNamespace(syscall.SysProcAttr{})
	containerCommand.SysProcAttr = &containerAttributes
	containerCommand.Stdout = os.Stdout
	containerCommand.Stdin = os.Stdin
	containerCommand.Stderr = os.Stderr
	return containerCommand.Run()
}
