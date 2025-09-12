package rpc

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"oci-runtime/internal/app"
	"oci-runtime/internal/infrastructure/technical/xerr"
)

type rpcRequest struct {
	ID      string
	OpName  string
	payload json.RawMessage
}

type rpcResponse struct {
	ReplyTo string
	Payload json.RawMessage
}

type rpcPipe struct {
	handlers map[string]app.RpcHandler
	ipc      app.Ipc
	a        string
}

func NewRpcPipe(ipc app.Ipc) *rpcPipe {
	return &rpcPipe{
		ipc: ipc,
	}
}

func (r *rpcPipe) id() string {
	b := make([]byte, 8) // 64 bits
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func (r *rpcPipe) Close() {
	r.ipc.Close()
}

func (r *rpcPipe) Op(name string, payload json.RawMessage) (json.RawMessage, error) {
	if err := r.ipc.Send(rpcRequest{
		ID:      r.id(),
		OpName:  name,
		payload: payload,
	}); err != nil {
		return nil, nil
	}

	var resp rpcResponse
	if err := r.ipc.Recv(&resp); err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (r *rpcPipe) Register(name string, handler app.RpcHandler) {
	r.handlers[name] = handler
}

func (r *rpcPipe) HandleOnce(ctx context.Context) error {
	var req rpcRequest
	if err := r.ipc.Recv(&req); err != nil {
		return err
	}
	h := r.handlers[req.OpName]
	if h == nil {
		return xerr.Op("no handler for that IPC request", nil, xerr.KV{"req": fmt.Sprintf("#+v", req)})
	}

	payload := h(ctx, req.payload)
	payloadJson, err := json.Marshal(payload)
	if err != nil {
		return nil
	}
	return r.ipc.Send(rpcResponse{
		ReplyTo: req.ID,
		Payload: payloadJson,
	})
}
