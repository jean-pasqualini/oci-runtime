package mount

import (
	"context"
	"golang.org/x/sys/unix"
	"oci-runtime/internal/domain"
	"oci-runtime/internal/infrastructure/technical/logging"
	"oci-runtime/internal/infrastructure/technical/xerr"
	"os"
)

type manager struct {
}

func NewManager() *manager {
	return &manager{}
}

func (m *manager) createTarget(ctx context.Context, mt Mount) error {
	l := logging.FromContext(ctx)
	if mt.FSType == "bind" {
		isSourceDir, err := isDir(mt.Source)
		if err != nil {
			return err
		}
		if isSourceDir {
			l.Debug("make directory")
			if err := os.MkdirAll(mt.Target, 0o555); err != nil {
				return xerr.Op("mkdir "+mt.Target, err, xerr.KV{})
			}
		} else {
			l.Debug("create file")
			if err := touch(mt.Target); err != nil {
				return err
			}
		}
	} else {
		l.Debug("make directory")
		if err := os.MkdirAll(mt.Target, 0o555); err != nil {
			return xerr.Op("mkdir "+mt.Target, err, xerr.KV{})
		}
	}

	return nil
}

func (m *manager) Mount(ctx context.Context, mc domain.ContainerMountConfiguration) error {
	l := logging.FromContext(ctx)
	mt := mapToMount(mc)
	l = l.With("params", mt)
	l.Debug("unmount")
	_ = unix.Unmount(mt.Target, 0)

	if err := m.createTarget(ctx, mt); err != nil {
		return err
	}

	l.Debug("mount")
	if err := unix.Mount(mt.Source, mt.Target, mt.FSType,
		mt.Flags, mt.Data); err != nil {
		return xerr.Op("mount "+mt.FSType, err, xerr.KV{})
	}
	return nil
}

func (m *manager) MakePrivate(ctx context.Context, path string) error {
	l := logging.FromContext(ctx)
	l.Debug("mount private", "path", path)
	if err := unix.Mount("", path, "", unix.MS_REC|unix.MS_PRIVATE, ""); err != nil {
		return xerr.Op("mount --make-rprivate "+path, err, xerr.KV{})
	}
	return nil
}
