package app

import (
	"context"
	"fmt"
	"golang.org/x/sys/unix"
	"oci-runtime/internal/app/mw"
	"oci-runtime/internal/platform/logging"
)

type CheckComamnd struct {
}

func NewCheckHandler() mw.HandlerFunc[CheckComamnd] {
	h := checkHandler{}
	return h.handle
}

type checkHandler struct {
}

func seccompModeString(mode int) string {
	switch mode {
	case unix.SECCOMP_MODE_DISABLED:
		return "disabled"
	case unix.SECCOMP_MODE_STRICT:
		return "strict"
	case unix.SECCOMP_MODE_FILTER:
		return "filter"
	default:
		return fmt.Sprintf("unknown (%d)", mode)
	}
}

// strace -f -e trace=%process,%file,%mount ./your-binary
func (h *checkHandler) handle(ctx context.Context, cmd CheckComamnd) error {
	logger := logging.FromContext(ctx)
	logger.Info("start checking")
	hdr := unix.CapUserHeader{
		Version: unix.LINUX_CAPABILITY_VERSION_3,
		Pid:     0,
	}

	var data unix.CapUserData

	if err := unix.Capget(&hdr, &data); err != nil {
		return nil
	}

	if data.Effective&(1<<unix.CAP_SYS_ADMIN) != 0 {
		fmt.Println("CAP_SYS_ADMIN effective ✅")
	}

	mode, _, errno := unix.Syscall6(unix.SYS_PRCTL, unix.PR_GET_SECCOMP, 0, 0, 0, 0, 0)
	if errno != 0 {
		return errno
	}

	fmt.Println(seccompModeString(int(mode)))

	return nil
}
