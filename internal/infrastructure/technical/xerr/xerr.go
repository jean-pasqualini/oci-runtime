// internal/xerr/xerr.go
package xerr

import (
	"log/slog"
)

type KV = map[string]string

type AttrError interface {
	error
	Attrs() map[string]string
}

type AttrLog interface {
	error
	LogAttrs() []any
}

type OpError struct {
	op    string
	err   error
	attrs map[string]string
}

func (e *OpError) Error() string            { return e.op + " : " + e.err.Error() }
func (e *OpError) Unwrap() error            { return e.err }
func (e *OpError) Attrs() map[string]string { return e.attrs }
func (e *OpError) LogAttrs() []any {
	la := make([]any, 0, len(e.attrs))
	for k, v := range e.attrs {
		la = append(la, slog.String(k, v))
	}
	if perr, ok := e.err.(AttrLog); ok {
		la = append(la, perr.LogAttrs()...)
	}

	return la
}

func Op(op string, err error, attrs map[string]string) error {
	if err == nil {
		return nil
	}
	return &OpError{op: op, err: err, attrs: attrs}
}
