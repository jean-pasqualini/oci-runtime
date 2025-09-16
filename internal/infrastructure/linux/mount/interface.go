package mount

import (
	"golang.org/x/sys/unix"
	"oci-runtime/internal/domain"
)

type Mount struct {
	Source string
	Target string
	FSType string
	Flags  uintptr
	Data   string
}

var mountOptionFlags = map[string]uintptr{
	"nosuid":     unix.MS_NOSUID,
	"nodev":      unix.MS_NODEV,
	"noexec":     unix.MS_NOEXEC,
	"ro":         unix.MS_RDONLY,
	"sync":       unix.MS_SYNCHRONOUS,
	"mand":       unix.MS_MANDLOCK,
	"dirsync":    unix.MS_DIRSYNC,
	"remount":    unix.MS_REMOUNT,
	"bind":       unix.MS_BIND,
	"rbind":      unix.MS_BIND | unix.MS_REC,
	"rec":        unix.MS_REC,
	"private":    unix.MS_PRIVATE,
	"shared":     unix.MS_SHARED,
	"slave":      unix.MS_SLAVE,
	"unbindable": unix.MS_UNBINDABLE,
}

func mapToMount(mc domain.ContainerMountConfiguration) Mount {
	var flags uintptr = 0
	for _, opt := range mc.Options {
		if f, ok := mountOptionFlags[opt]; ok {
			flags |= f
		}
	}

	if flags&unix.MS_BIND != 0 {
		// Easy to ensure we have always that type
		// Even if the kernel doesn't care of the type for a bind
		mc.Type = "bind"
	}

	return Mount{
		Source: mc.Source,
		Target: mc.Destination,
		FSType: mc.Type,
		Flags:  flags,
		Data:   "",
	}
}
