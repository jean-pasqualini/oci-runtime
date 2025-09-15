package app

import (
	"context"
	"encoding/json"
	"io"
	"oci-runtime/internal/domain"
)

type IpcFactory func(rPipe io.Reader, wPipe io.Writer) Ipc
type Ipc interface {
	Send(ctx context.Context, data any) error
	Recv(ctx context.Context, v any) error
	Close()
}

type RpcHandler func(ctx context.Context, payload json.RawMessage) json.RawMessage
type Rpc interface {
	Op(ctx context.Context, name string, payload json.RawMessage) (json.RawMessage, error)
	Register(name string, handler RpcHandler)
	HandleOnce(ctx context.Context) error
	Close()
}

type NetworkManager interface {
	BringUp(ctx context.Context, name string) error
	AddAddr(ctx context.Context, ipCIDR string) error
}
type MountManager interface {
	MakePrivate(ctx context.Context, path string) error
	Mount(ctx context.Context, mt domain.ContainerMountConfiguration) error
}

type NamespaceManager interface {
	SetHostname(ctx context.Context, name string) error
}

type RootSwitcher interface {
	Chroot(ctx context.Context, containerRoot string) error
	Pivot(ctx context.Context, containerRoot string) error
}

type Process interface {
	SetComm(ctx context.Context, name string) error
	Exec(ctx context.Context, argv []string, env []string) error
}
