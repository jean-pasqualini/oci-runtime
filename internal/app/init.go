package app

import (
	"context"
	"fmt"
	"golang.org/x/sys/unix"
	"oci-runtime/internal/app/mw"
	"oci-runtime/internal/domain"
	"oci-runtime/internal/infrastructure/technical/logging"
	"oci-runtime/internal/infrastructure/technical/xerr"
	"os"
	"strconv"
)

type InitCmd struct{}

type Ports struct {
	Mount      MountManager
	NS         NamespaceManager
	Root       RootSwitcher
	Proc       Process
	Net        NetworkManager
	IpcFactory IpcFactory
}

func NewInitHandler(p Ports) mw.HandlerFunc[InitCmd] {
	h := initHandler{p}
	return h.handle
}

type initHandler struct {
	p Ports
}

func (h *initHandler) rootChroot(ctx context.Context, containerRoot string) error {
	l := logging.FromContext(ctx)
	l.Info("root chroot", "directory", containerRoot)

	if err := h.p.Root.Chroot(ctx, containerRoot); err != nil {
		return err
	}

	return nil
}

func (h *initHandler) rootPivot(ctx context.Context, containerRoot string) error {
	l := logging.FromContext(ctx)
	l.Info("root pivot", "container_root", containerRoot)
	if err := h.p.Root.Pivot(ctx, containerRoot); err != nil {
		return err
	}

	return nil
}

func (h *initHandler) switchRoot(ctx context.Context, containerRoot string) error {
	l := logging.FromContext(ctx)
	if err := h.rootPivot(ctx, containerRoot); err != nil {
		l.Warn("change root (pivot) failed : ", "err", err.Error())
		if err := h.rootChroot(ctx, containerRoot); err != nil {
			return xerr.Op("chroot", err, xerr.KV{})
		}
	}

	return nil
}

func (h *initHandler) prepareProcess(ctx context.Context) error {
	_ = h.p.Proc.SetComm(ctx, "oci-rn:[1:INIT]")
	return h.p.NS.SetHostname(ctx, "matrix-container")
}

func (h *initHandler) configureIsolation(ctx context.Context, containerRoot string) error {
	if err := h.p.Mount.MakePrivate(ctx, "/"); err != nil {
		return xerr.Op("make private /", err, xerr.KV{})
	}
	if err := h.switchRoot(ctx, containerRoot); err != nil {
		return xerr.Op("enter root", err, xerr.KV{})
	}

	return nil
}

func (h *initHandler) setupFilesystem(ctx context.Context) error {
	l := logging.FromContext(ctx)
	l.Info("setup mount point /proc")
	if err := h.p.Mount.Mount(ctx, domain.Mount{
		Source: "proc",
		Target: "/proc",
		FSType: "proc",
		Flags:  uintptr(unix.MS_NOSUID | unix.MS_NODEV | unix.MS_NOEXEC),
	}); err != nil {
		return xerr.Op("mount proc", err, xerr.KV{})
	}

	return nil
}

func (h *initHandler) launchEntrypoint(ctx context.Context, argv []string) error {
	env := os.Environ()
	return h.p.Proc.Exec(ctx, argv, env)
}

func (h *initHandler) configureNetwork(ctx context.Context) error {
	l := logging.FromContext(ctx)
	l.Info("configure network: up lo")
	if err := h.p.Net.BringUp(ctx, "lo"); err != nil {
		return err
	}
	l.Info("configure network: add a testing ip v4 on lo")
	if err := h.p.Net.AddAddr(ctx, "192.168.1.0/24"); err != nil {
		return err
	}
	l.Info("configure network: add a testing ip v6 on lo")
	if err := h.p.Net.AddAddr(ctx, "fd00::1/128"); err != nil {
		return nil
	}

	return nil
}

func (h *initHandler) closeExtraFDs() error {
	// ^uint(0) equals to the not binary, it inverts all the bit so we obtains the max value
	// equals to UINT_MAX in C
	if err := unix.CloseRange(3, 100, unix.CLOSE_RANGE_CLOEXEC); err != nil {
		return err
	}

	return nil
}

func (h *initHandler) handle(ctx context.Context, _ InitCmd) error {
	l := logging.FromContext(ctx)

	l.Info("start")

	sPipeReadEnv, _ := strconv.Atoi(os.Getenv("FD_SYNC_READ"))
	sPipeReadFD := os.NewFile(uintptr(sPipeReadEnv), "sync-pipe-read")
	sPipeWriteEnv, _ := strconv.Atoi(os.Getenv("FD_SYNC_WRITE"))
	sPipeWriteFD := os.NewFile(uintptr(sPipeWriteEnv), "sync-pipe-write")
	syncIpc := h.p.IpcFactory(sPipeReadFD, sPipeWriteFD)

	var containerConfig domain.ContainerConfiguration
	if err := syncIpc.Recv(&containerConfig); err != nil {
		return err
	}

	if err := h.prepareProcess(ctx); err != nil {
		l.Warn("hostname set failed", "error", err)
	}
	if err := h.configureIsolation(ctx, containerConfig.Root.Path); err != nil {
		return err
	}
	if err := h.setupFilesystem(ctx); err != nil {
		return xerr.Op("setup filesystem", err, xerr.KV{})
	}
	if err := h.configureNetwork(ctx); err != nil {
		return xerr.Op("configure network", err, xerr.KV{})
	}
	var initDone bool
	err := syncIpc.Send(&initDone)
	if err != nil {
		return err
	}

	ePipeReadEnv, err := strconv.Atoi(os.Getenv("FD_EXEC"))
	if err != nil {
		return err
	}
	l.Info("waiting start order", "execPipe", ePipeReadEnv)
	//ePipeReadFD, err := os.OpenFile(fmt.Sprintf("/proc/self/fd/%d", ePipeReadEnv), unix.O_RDONLY|unix.O_CLOEXEC, 0)
	ePipeReadFD, err := os.OpenFile(fmt.Sprintf("/proc/self/fd/%d", ePipeReadEnv), os.O_RDONLY, 0)
	//ePipeReadFD := os.NewFile(uintptr(ePipeReadEnv), "exec.pipe")
	execIpc := h.p.IpcFactory(ePipeReadFD, nil)

	var waitingStart bool
	if err := execIpc.Recv(&waitingStart); err != nil {
		return err
	}
	l.Info("start order received, run entrypoint")

	// Ensure all FD are closed to avoid leak security leak to the entry point
	if err := h.closeExtraFDs(); err != nil {
		return err
	}

	if err := h.launchEntrypoint(ctx, containerConfig.Process.Args); err != nil {
		return xerr.Op("run main process", err, xerr.KV{})
	}

	return nil
}
