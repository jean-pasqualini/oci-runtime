//go:build medium

package network

import (
	"context"
	"errors"
	"fmt"
	"github.com/mdlayher/netlink"
	"golang.org/x/sys/unix"
	"net"
	"oci-runtime/internal/infrastructure/technical/logging"
)

func (m *manager) BringUp(ctx context.Context, name string) error {
	ifi, err := net.InterfaceByName(name)
	if err != nil {
		return fmt.Errorf("iface %q: %w", name, err)
	}

	// Open a NETLINK_ROUTE connection.
	conn, err := netlink.Dial(unix.NETLINK_ROUTE, nil)
	if err != nil {
		return fmt.Errorf("netlink dial: %w", err)
	}
	defer conn.Close()

	// Compose the netlink message: RTM_NEWLINK + REQUEST|ACK
	msg := netlink.Message{
		Header: netlink.Header{
			Type:  unix.RTM_NEWLINK,
			Flags: netlink.Request | netlink.Acknowledge,
		},
		Data: structToBytes(createIfupMessage(int32(ifi.Index))),
	}

	// Execute and expect an ACK (NLMSG_ERROR with code 0).
	_, err = conn.Execute(msg)
	if err != nil {
		return fmt.Errorf("RTM_NEWLINK IFF_UP: %w", err)
	}
	return nil
}

func (m *manager) createNetlinkMessage(data []byte) netlink.Message {
	return netlink.Message{
		Header: netlink.Header{
			Type:  unix.RTM_NEWADDR,
			Flags: netlink.Request | netlink.Acknowledge | netlink.Create | netlink.Excl,
		},
		Data: data,
	}
}

func (m *manager) createAttributes(family int, ip net.IP) []netlink.Attribute {
	if family == unix.AF_INET {
		return []netlink.Attribute{
			{Type: unix.IFA_LOCAL, Data: ip.To4()},
			{Type: unix.IFA_ADDRESS, Data: ip.To4()},
		}
	}
	return []netlink.Attribute{
		{Type: unix.IFA_ADDRESS, Data: ip.To16()},
	}
}

func (m *manager) AddAddr(ctx context.Context, ipCIDR string) error {
	l := logging.FromContext(ctx)
	// Parse CIDR
	ip, ipNet, err := net.ParseCIDR(ipCIDR)
	if err != nil {
		return nil
	}

	l.Debug("parse CIDR", "ip", ipNet.IP, "mask", ipNet.Mask)

	ifi, err := net.InterfaceByName("lo")
	if err != nil {
		return err
	}
	l.Debug("found interface", "name", ifi.Name)

	// Open a NETLINK_ROUTE connection.
	conn, err := netlink.Dial(unix.NETLINK_ROUTE, nil)
	if err != nil {
		return fmt.Errorf("netlink dial: %w", err)
	}
	defer conn.Close()

	family, err := determineIPFamily(ip)
	if err != nil {
		return nil
	}

	l.Debug("found family", "family", family)

	// Fast idempotency check: list existing addresses on the iface, same family.
	// Add address
	attrs := m.createAttributes(family, ip)
	attrData, err := netlink.MarshalAttributes(attrs)
	if err != nil {
		return nil
	}
	prefixLen, _ := ipNet.Mask.Size()
	ifam := createIfAddrmsg(family, prefixLen, ifi.Index)
	msg := m.createNetlinkMessage(append(structToBytes(ifam), attrData...))

	_, err = conn.Execute(msg)
	if err != nil {
		var oe *netlink.OpError
		if errors.As(err, &oe) && oe.Err == unix.EEXIST {
			l.Debug("already existing ip", "ip", ip.String())
			return nil
		}
		return err
	}
	return nil
}
