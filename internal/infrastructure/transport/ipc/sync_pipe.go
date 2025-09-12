package ipc

import (
	"encoding/json"
	"io"
	"os"
)

type syncPipe struct {
	rPipe io.Reader
	wPipe io.Writer
}

func NewSyncPipe(rPipe io.Reader, wPipe io.Writer) *syncPipe {
	return &syncPipe{
		rPipe,
		wPipe,
	}
}

func (s *syncPipe) Close() {
	s.rPipe.(*os.File).Close()
	s.wPipe.(*os.File).Close()
}

func (s *syncPipe) Send(data any) error {
	enc := json.NewEncoder(s.wPipe)
	if err := enc.Encode(data); err != nil {
		return err
	}

	return nil
}

func (s *syncPipe) Recv(data any) error {
	dec := json.NewDecoder(s.rPipe)
	if err := dec.Decode(data); err != nil {
		return err
	}

	return nil
}
