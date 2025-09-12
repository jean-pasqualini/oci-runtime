//go:build easy

package network

import (
	"context"
	"fmt"
	"github.com/vishvananda/netlink"
)

func (m *manager) BringUp(ctx context.Context, name string) error {
	link, err := netlink.LinkByName(name)
	if err != nil {
		return fmt.Errorf("get link %s: %w", name, err)
	}
	if err := netlink.LinkSetUp(link); err != nil {
		return fmt.Errorf("set link %s up: %w", name, err)
	}
	return nil
}

func (m *manager) AddAddr(ctx context.Context, ipCIDR string) error {
	link, err := netlink.LinkByName(name)
	if err != nil {
		return fmt.Errorf("get link %s: %w", name, err)
	}

	addr, err := netlink.ParseAddr(ipCIDR)
	if err != nil {
		return err
	}

	if err := netlink.AddrAdd(link, nil); err != nil {
		return err
	}

	netlink.LinkSet
	return nil
}
