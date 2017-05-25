package wol

import (
	"bytes"
	"fmt"
	"io"
	"net"
)

type MagicPacket []byte

// Create a magic packet for the given hwAddr.
func NewMagicPacket(hwAddr net.HardwareAddr) MagicPacket {
	const hwAddrN = 16
	bcastAddr := []byte{255, 255, 255, 255, 255, 255}
	off := len(bcastAddr)
	p := make([]byte, off+(hwAddrN*len(hwAddr)))
	copy(p, bcastAddr)
	copy(p[off:], bytes.Repeat(hwAddr, hwAddrN))
	return p
}

// Wake sends a magic packet for hwAddr to the broadcast address. If src is not nil, it is used as the local address for
// the broadcast.
func Wake(src net.IP, hwAddr net.HardwareAddr) error {
	var laddr *net.UDPAddr
	if src != nil {
		laddr = &net.UDPAddr{IP: src}
	}
	raddr := &net.UDPAddr{IP: net.IPv4bcast, Port: 9}
	conn, err := net.DialUDP("udp", laddr, raddr)
	if err != nil {
		return err
	}
	p := NewMagicPacket(hwAddr)
	n, err := conn.Write([]byte(p))
	if err == nil && n < len(p) {
		return io.ErrShortWrite
	}
	if err1 := conn.Close(); err1 != nil {
		err = err1
	}
	return err
}

// WakeString sends a magic packet for macAddr to the broadcast address. If srcIP non-empty, it is used as the local
// address for the broadcast.
func WakeString(srcIP, macAddr string) error {
	hwAddr, err := net.ParseMAC(macAddr)
	if err != nil {
		return err
	}
	var src net.IP
	if srcIP != "" {
		src = net.ParseIP(srcIP)
		if src == nil {
			return fmt.Errorf("invalid ip: %s", srcIP)
		}
	}
	return Wake(src, hwAddr)
}
