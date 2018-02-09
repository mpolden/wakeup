package wol

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"sync"
)

// Bridge represents a Wake-on-LAN bridge.
type Bridge struct {
	conn     io.ReadCloser
	lastSent MagicPacket
	wakeFunc func(net.IP, net.HardwareAddr) error
	mu       sync.Mutex
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

// Close closes the connection.
func Close(b *Bridge) error { return b.conn.Close() }

// Forward reads a magic packet and writes it back to the network using src as the local address.
func (b *Bridge) Forward(src net.IP) (MagicPacket, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	mp, err := b.read()
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

func (b *Bridge) read() (MagicPacket, error) {
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
