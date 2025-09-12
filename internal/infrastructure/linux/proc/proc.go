package proc

import (
	"context"
	"golang.org/x/sys/unix"
	"oci-runtime/internal/infrastructure/technical/logging"
	"oci-runtime/internal/infrastructure/technical/xerr"
	"strings"
	"syscall"
	"unsafe"
)

type manager struct {
}

func NewManager() *manager {
	return &manager{}
}

func (m *manager) SetComm(ctx context.Context, name string) error {
	l := logging.FromContext(ctx)
	l.Info("set comm", "name", name)
	b := make([]byte, 16) // 15 + NUL
	copy(b, name)
	if err := unix.Prctl(syscall.PR_SET_NAME, uintptr(unsafe.Pointer(&b[0])), 0, 0, 0); err != nil {
		return xerr.Op("prctl(PR_SET_NAME)", err, xerr.KV{
			"name": name,
		})
	}
	return nil
}

func (m *manager) Exec(ctx context.Context, argv []string, env []string) error {
	l := logging.FromContext(ctx)
	l.Info("execve", "argv", strings.Join(argv, " "))
	if err := unix.Exec(argv[0], argv, env); err != nil {
		return xerr.Op("execve", err, xerr.KV{})
	}

	return nil
}
