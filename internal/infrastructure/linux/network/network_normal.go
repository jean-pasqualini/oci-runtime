//go:build normal

package network

import (
	"context"
	"github.com/jsimonetti/rtnetlink"
	"golang.org/x/sys/unix"
	"log"
	"net"
	"oci-runtime/internal/infrastructure/technical/logging"
	"oci-runtime/internal/infrastructure/technical/xerr"
	"time"
)

func (m *manager) BringUp(ctx context.Context, name string) error {
	l := logging.FromContext(ctx)
	l = l.With("name", name)
	ifi, err := net.InterfaceByName(name)
	if err != nil {
		log.Fatalf("failed to lookup interface: %v", err)
	}

	c, err := rtnetlink.Dial(nil)
	if err != nil {
		return err
	}
	defer c.Close()

	link, err := c.Link.Get(uint32(ifi.Index))
	if err != nil {
		return xerr.Op("get interface", err, xerr.KV{})
	}

	link.Flags = unix.IFF_UP
	link.Change = unix.IFF_UP
	link.Attributes = nil

	if err := c.Link.Set(&link); err != nil {
		return xerr.Op("set interface up", err, xerr.KV{})
	}

	return nil
}

func (m *manager) SetIp() error {
	// Dial a rtnetlink connection
	conn, err := rtnetlink.Dial(nil)
	if err != nil {
		log.Fatalf("failed to dial rtnetlink: %v", err)
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Find interface index
	ifaceName := "eth0"
	iface, err := net.InterfaceByName(ifaceName)
	if err != nil {
		log.Fatalf("failed to get interface: %v", err)
	}

	// IP and mask to add
	ipNet := &net.IP{
		IP:   net.ParseIP("192.168.1.100"),
		Mask: net.CIDRMask(24, 32), // 192.168.1.0/24
	}

	msg := rtnetlink.AddressMessage{
		Attributes: &rtnetlink.AddressAttributes{
			Address: &ipNet,
		},
	}

	// Add address to interface
	err = conn.Address.New(
		ctx,
		iface.Index,
		ipNet,
	)
	if err != nil {
		log.Fatalf("failed to add IP: %v", err)
	}

	log.Printf("Added %s to %s", ipNet.String(), ifaceName)
}
