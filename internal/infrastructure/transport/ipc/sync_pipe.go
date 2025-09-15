package ipc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"oci-runtime/internal/app"
	"oci-runtime/internal/infrastructure/technical/logging"
	"oci-runtime/internal/infrastructure/technical/xerr"
	"os"
)

type syncPipe struct {
	rPipe io.Reader
	wPipe io.Writer
}

func NewSyncPipe(rPipe io.Reader, wPipe io.Writer) app.Ipc {
	return &syncPipe{
		rPipe,
		wPipe,
	}
}

func (s *syncPipe) Close() {
	if s.rPipe != nil {
		s.rPipe.(*os.File).Close()
	}
	if s.wPipe != nil {
		s.wPipe.(*os.File).Close()
	}
}

func (s *syncPipe) Send(ctx context.Context, data any) error {
	l := logging.FromContext(ctx)
	if s.wPipe == nil {
		return fmt.Errorf("no pipe write, can't send data")
	}
	var dbg bytes.Buffer
	mw := io.MultiWriter(s.wPipe, &dbg)
	enc := json.NewEncoder(mw)
	if err := enc.Encode(data); err != nil {
		return xerr.Op("Send", err, xerr.KV{})
	}

	l.Debug("send ipc packet", "content", dbg.String())

	return nil
}

func (s *syncPipe) Recv(_ context.Context, data any) error {
	if s.rPipe == nil {
		return fmt.Errorf("no pipe read, can't read data")
	}

	var buf bytes.Buffer
	tee := io.TeeReader(s.rPipe, &buf)
	dec := json.NewDecoder(tee)
	if err := dec.Decode(data); err != nil {
		return xerr.Op("Recv decode error", err, xerr.KV{"b": buf.String()})
	}

	return nil
}
