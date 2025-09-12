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

func (m *manager) Mount(ctx context.Context, mt domain.Mount) error {
	l := logging.FromContext(ctx)
	l = l.With("params", mt)
	l.Debug("unmount")
	_ = unix.Unmount(mt.Target, 0)
	l.Debug("make directory")
	if err := os.MkdirAll(mt.Target, 0o555); err != nil {
		return xerr.Op("mkdir "+mt.Target, err, xerr.KV{})
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
