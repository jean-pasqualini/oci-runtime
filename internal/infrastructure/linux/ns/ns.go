package ns

import (
	"context"
	"golang.org/x/sys/unix"
	"oci-runtime/internal/infrastructure/technical/xerr"
)

type manager struct {
}

func NewManager() *manager {
	return &manager{}
}

func (m *manager) SetHostname(ctx context.Context, name string) error {
	if err := unix.Sethostname([]byte(name)); err != nil {
		// Wrap technique, avec détails utiles pour le debug
		return xerr.Op("ns.sethostname", err, xerr.KV{"hostname": name})
	}
	return nil
}
