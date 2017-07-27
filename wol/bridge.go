package wol

import (
	"bytes"
	"fmt"
	"io"
	"net"
)

// Bridge represents a Wake-on-LAN bridge.
type Bridge struct {
	conn     io.Reader
	lastSent MagicPacket
	wakeFunc func(net.IP, net.HardwareAddr) error
}

// Listen listens for magic packets on the given addr.
func Listen(addr string) (*Bridge, error) {
	udpAddr, err := net.ResolveUDPAddr("udp4", addr)
	if err != nil {
		return nil, err
	}
	conn, err := net.ListenUDP("udp4", udpAddr)
	if err != nil {
		return nil, err
	}
	return &Bridge{conn: conn, wakeFunc: Wake}, nil
}

// ReadMagicPacket reads magic packets using the bridge
func (b *Bridge) ReadMagicPacket() (MagicPacket, error) {
	buf := make([]byte, 4096)
	n, err := b.conn.Read(buf)
	if err != nil {
		return nil, err
	}
	mp := buf[:n]
	if !IsMagicPacket(mp) {
		return nil, fmt.Errorf("invalid magic packet: %x", mp)
	}
	return mp, nil
}

// Forward reads a magic packet and writes to the bridged network using src as the local address.
func (b *Bridge) Forward(src net.IP) (MagicPacket, error) {
	mp, err := b.ReadMagicPacket()
	if err != nil {
		return nil, err
	}
	// Do not resend if we just sent this packet
	if bytes.Equal(mp, b.lastSent) {
		b.lastSent = nil
		return nil, nil
	}
	if err := b.wakeFunc(src, mp.HardwareAddr()); err != nil {
		return nil, err
	}
	b.lastSent = mp
	return mp, nil
}
