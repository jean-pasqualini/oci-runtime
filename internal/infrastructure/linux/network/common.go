package network

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"golang.org/x/sys/unix"
	"net"
	"oci-runtime/internal/infrastructure/technical/xerr"
)

// Build an ifinfomsg payload (16 bytes on 64-bit):
//
//	struct ifinfomsg {
//	  __u8  ifi_family;
//	  __u8  __ifi_pad;
//	  __u16 ifi_type;
//	  __s32 ifi_index;
//	  __u32 ifi_flags;
//	  __u32 ifi_change;
//	};
func createIfupMessage(index int32) unix.IfInfomsg {
	return unix.IfInfomsg{
		Family: unix.AF_UNSPEC,
		Index:  index,
		Flags:  unix.IFF_UP,
		Change: unix.IFF_UP,
	}
}

func createIfAddrmsg(family int, prefixLen int, index int) unix.IfAddrmsg {
	return unix.IfAddrmsg{
		Family:    byte(family),
		Prefixlen: byte(prefixLen),
		Flags:     0,
		Scope:     unix.RT_SCOPE_UNIVERSE,
		Index:     uint32(index),
	}
}

func determineIPFamily(ip net.IP) (int, error) {
	// Determine family
	family := unix.AF_UNSPEC
	if ip.To4() != nil {
		family = unix.AF_INET
	} else if ip.To16() != nil {
		family = unix.AF_INET6
	} else {
		return 0, xerr.Op(
			"determine family",
			fmt.Errorf("not an ip v4"),
			xerr.KV{"family": string(rune(family))},
		)
	}

	return family, nil
}

func structToBytes(s interface{}) []byte {
	var buf bytes.Buffer
	if err := binary.Write(&buf, binary.LittleEndian, s); err != nil {
		return []byte("")
	}

	return buf.Bytes()
}

type manager struct {
}

func NewManager() *manager {
	return &manager{}
}
