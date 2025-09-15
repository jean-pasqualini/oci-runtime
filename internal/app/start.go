package app

import (
	"context"
	"oci-runtime/internal/app/mw"
	"oci-runtime/internal/infrastructure/technical/logging"
	"os"
	"path"
	"path/filepath"
)

type StartCmd struct {
	Name         string
	MetadataRoot string
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
	containerStateFolder := filepath.Join(cmd.MetadataRoot, cmd.Name)

	ePipeWritePath := path.Join(containerStateFolder, "exec.fifo")

	ePipeWriteFD, err := os.OpenFile(ePipeWritePath, os.O_WRONLY, 0)
	if err != nil {
		return err
	}
	execPipe := h.ipcFactory(nil, ePipeWriteFD)
	defer execPipe.Close()
	var giveStartOrder bool
	if err := execPipe.Send(ctx, &giveStartOrder); err != nil {
		return err
	}

	logger.Debug("start order given")
	return nil
}
