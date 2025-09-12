package rs

import (
	"context"
	"golang.org/x/sys/unix"
	"oci-runtime/internal/infrastructure/technical/xerr"
	"os"
	"path/filepath"
)

type manager struct {
}

const fileModeOwnerOnly = os.FileMode(0o700)

func NewManager() *manager {
	return &manager{}
}

func (h *manager) Chroot(ctx context.Context, containerRoot string) error {
	if err := unix.Chroot(containerRoot); err != nil {
		return err
	}
	if err := unix.Chdir("/"); err != nil {
		return err
	}

	return nil
}

func (h *manager) Pivot(ctx context.Context, containerRoot string) error {
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
