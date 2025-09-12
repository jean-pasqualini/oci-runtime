//go:build advanced

package network

import (
	"context"
	"encoding/binary"
	"fmt"
	"golang.org/x/sys/unix"
	"net"
	"oci-runtime/internal/infrastructure/technical/logging"
	"oci-runtime/internal/infrastructure/technical/xerr"
	"unsafe"
)

func nlMsg(typ, flags int, payload []byte) []byte {
	hlen := unix.NLMSG_HDRLEN
	plen := len(payload)
	b := make([]byte, hlen+plen)
	h := unix.NlMsghdr{
		Len:   uint32(hlen + plen),
		Type:  uint16(typ),
		Flags: uint16(flags),
		Seq:   1,
		Pid:   0,
	}

	copy(b[0:], structToBytes(h))
	copy(b[hlen:], payload)
	return b
}

func nlSend(fd int, msg []byte) error {
	return unix.Sendto(fd, msg, 0, &unix.SockaddrNetlink{
		Family: unix.AF_NETLINK,
	})
}

func nlRecvAck(fd int) error {
	buf := make([]byte, 8192)
	n, _, err := unix.Recvfrom(fd, buf, 0)
	if err != nil {
		return fmt.Errorf("recv: %w", err)
	}
	// It is possible to receive multiple concatenated messages
	p := buf[:n]
	// As long as we have at least a header
	for len(p) >= unix.NLMSG_HDRLEN {
		// We extract the header
		h := (*unix.NlMsghdr)(unsafe.Pointer(&p[0]))
		// Len should have the perfect size
		if int(h.Len) < unix.NLMSG_HDRLEN || int(h.Len) > len(p) {
			return fmt.Errorf("bad nlmsg len")
		}
		// We extract teh message part
		msg := p[unix.NLMSG_HDRLEN:int(h.Len)]
		// If that's an error message
		if h.Type == unix.NLMSG_ERROR {
			// The message should by 4 bytes
			if len(msg) < 4 {
				return fmt.Errorf("short NLMSG_ERROR")
			}
			code := int32(binary.LittleEndian.Uint32(msg[:4]))
			// When code is 0, all is fine
			if code == 0 {
				return nil
			} // ACK OK
			return fmt.Errorf("netlink error: %d", code)
		}
		// next (4-byte aligned)
		adv := (int(h.Len) + unix.NLMSG_ALIGNTO - 1) & ^(unix.NLMSG_ALIGNTO - 1)
		if adv <= 0 || adv > len(p) {
			break
		}
		p = p[adv:]
	}
	return nil
}

func (m *manager) BringUp(ctx context.Context, name string) error {
	l := logging.FromContext(ctx)
	l = l.With("name", name)

	l.Debug("lookup interface")
	ifi, err := net.InterfaceByName(name)
	if err != nil {
		return xerr.Op("lookup interface", err, xerr.KV{
			"dev": name,
		})
	}

	l.Debug("open a netlink socket", "ifi", ifi.Name)
	fd, err := unix.Socket(unix.AF_NETLINK, unix.SOCK_RAW, unix.NETLINK_ROUTE)
	if err != nil {
		return xerr.Op("no idea", err, xerr.KV{
			"dev": name,
		})
	}
	defer unix.Close(fd)

	ifm := createIfupMessage(ifi.Index)

	msg := nlMsg(unix.RTM_NEWLINK, unix.NLM_F_REQUEST|unix.NLM_F_ACK, structToBytes(ifm))
	l.Debug("send a message to netlink", "ifi", ifi.Name, "content", ifm)
	if err := nlSend(fd, msg); err != nil {
		return err
	}
	return nlRecvAck(fd)
}
