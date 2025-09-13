package app

import (
	"context"
	"oci-runtime/internal/app/mw"
	"oci-runtime/internal/infrastructure/technical/logging"
	"os"
	"path"
)

type StartCmd struct {
	Name      string
	StatePath string
}

func NewStartHandler(ipcFactory IpcFactory) mw.HandlerFunc[StartCmd] {
	h := startHandler{ipcFactory: ipcFactory}
	return h.handle
}

type startHandler struct {
	ipcFactory IpcFactory
}

func (h *startHandler) handle(ctx context.Context, cmd StartCmd) error {
	logger := logging.FromContext(ctx)
	logger.Info("oci-runtime start")
	ePipeWritePath := path.Join(cmd.StatePath, "exec.fifo")

	ePipeWriteFD, err := os.OpenFile(ePipeWritePath, os.O_WRONLY, 0)
	if err != nil {
		return err
	}
	execPipe := h.ipcFactory(nil, ePipeWriteFD)
	defer execPipe.Close()
	var giveStartOrder bool
	if err := execPipe.Send(&giveStartOrder); err != nil {
		return err
	}

	logger.Info("start order given")

	logger.Info("oci runtime finished")
	return nil
}
