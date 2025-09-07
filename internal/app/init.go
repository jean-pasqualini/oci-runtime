package app

import (
	"context"
	"log/slog"
	"oci-runtime/internal/platform/xerr"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/unix"
	"oci-runtime/internal/app/mw"
	"oci-runtime/internal/platform/logging"
)

type InitCmd struct{}

const fileModeOwnerOnly = os.FileMode(0o700)

func NewInitHandler() mw.HandlerFunc[InitCmd] {
	h := initHandler{}
	return h.handle
}

type initHandler struct{}

func (h *initHandler) setInitComm(ctx context.Context) error {
	l := logging.FromContext(ctx)
	l.Debug("set process comm")
	name := "oci-rn:[1:INIT]"
	b := make([]byte, 16) // 15 + NUL
	copy(b, name)
	if err := unix.Prctl(syscall.PR_SET_NAME, uintptr(unsafe.Pointer(&b[0])), 0, 0, 0); err != nil {
		return xerr.Op("prctl(PR_SET_NAME)", err, xerr.KV{
			"name": name,
		})
	}
	return nil
}

func (h *initHandler) setHostname(ctx context.Context, hostname string) error {
	l := logging.FromContext(ctx)
	l.Info("set hostname", "hostname", hostname)
	if err := unix.Sethostname([]byte(hostname)); err != nil {
		return xerr.Op("sethostname", err, xerr.KV{})
	}
	return nil
}

func (h *initHandler) makeRootPrivate(ctx context.Context) error {
	l := logging.FromContext(ctx)
	l.Info("disable mount propagation")
	if err := unix.Mount("", "/", "", unix.MS_REC|unix.MS_PRIVATE, ""); err != nil {
		return xerr.Op("mount --make-rprivate /", err, xerr.KV{})
	}
	return nil
}

func (h *initHandler) enterChroot(ctx context.Context) error {
	l := logging.FromContext(ctx)
	pwd, _ := os.Getwd()
	chrootDirectory := pwd + "/root"

	l.Info("enter chroot", "directory", chrootDirectory)

	if err := unix.Chroot(chrootDirectory); err != nil {
		return err
	}
	if err := unix.Chdir("/"); err != nil {
		return err
	}

	return nil
}

func (h *initHandler) pivotRoot(ctx context.Context, containerRoot string) error {
	l := logging.FromContext(ctx)
	l.Info("pivot root", "container_root", containerRoot)
	putOld := filepath.Join(containerRoot, ".oldroot")

	// Ensure absRoot is a mount point (self-bind).
	if err := unix.Mount(containerRoot, containerRoot, "", unix.MS_BIND|unix.MS_REC, ""); err != nil {
		return xerr.Op("mount --bind newRoot newRoot", err, xerr.KV{})
	}
	if err := os.MkdirAll(putOld, fileModeOwnerOnly); err != nil {
		return xerr.Op("mkdir .oldroot", err, xerr.KV{})
	}
	if err := unix.PivotRoot(containerRoot, putOld); err != nil {
		return xerr.Op("pivot_root", err, xerr.KV{
			"new_root": containerRoot,
		})
	}
	if err := os.Chdir("/"); err != nil {
		return xerr.Op("chdir(/)", err, xerr.KV{})
	}
	// Drop old root (detach so fds/cwd don't block).
	if err := unix.Unmount("/.oldroot", unix.MNT_DETACH); err != nil {
		return xerr.Op("umount /.oldroot", err, xerr.KV{})
	}
	_ = os.Remove("/.oldroot")
	return nil
}

func (h *initHandler) switchRoot(ctx context.Context) error {
	l := logging.FromContext(ctx)
	// Compute container root relative to current cwd
	pwd, err := os.Getwd()
	if err != nil {
		return xerr.Op("getwd", err, xerr.KV{})
	}
	containerRoot := filepath.Join(pwd, "root")

	if err := h.pivotRoot(ctx, containerRoot); err != nil {
		l.Warn("change root (pivot) failed : ", "err", err.Error())

		if err := h.enterChroot(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (h *initHandler) remountProc(ctx context.Context) error {
	l := logging.FromContext(ctx)
	l.Info("remount /proc")
	_ = unix.Unmount("/proc", 0)
	if err := os.MkdirAll("/proc", 0o555); err != nil {
		return xerr.Op("mkdir /proc", err, xerr.KV{})
	}
	if err := unix.Mount("proc", "/proc", "proc",
		uintptr(unix.MS_NOSUID|unix.MS_NODEV|unix.MS_NOEXEC), ""); err != nil {
		return xerr.Op("mount procfs", err, xerr.KV{})
	}
	return nil
}

func (h *initHandler) runMainProcess(ctx context.Context) error {
	l := logging.FromContext(ctx)
	argv := []string{"/bin/bash", "-l"} // argv[0] present
	env := os.Environ()
	l.Info("execve", "argv", strings.Join(argv, " "))

	if err := unix.Exec(argv[0], argv, env); err != nil {
		return xerr.Op("execve", err, xerr.KV{})
	}

	return nil
}

func (h *initHandler) handle(ctx context.Context, _ InitCmd) error {
	l := logging.FromContext(ctx)
	if l == nil {
		l = slog.Default()
	}

	l.Info("start")

	if err := h.setInitComm(ctx); err != nil {
		l.Warn("process comm set failed", "error", err)
	}
	if err := h.setHostname(ctx, "matrix-container"); err != nil {
		l.Warn("hostname set failed", "error", err)
	}
	if err := h.makeRootPrivate(ctx); err != nil {
		return err
	}
	if err := h.switchRoot(ctx); err != nil {
		return xerr.Op("switch root", err, xerr.KV{})
	}
	if err := h.remountProc(ctx); err != nil {
		return xerr.Op("remount proc", err, xerr.KV{})
	}
	if err := h.runMainProcess(ctx); err != nil {
		return xerr.Op("run main process", err, xerr.KV{})
	}

	return nil
}
