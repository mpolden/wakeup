package wol

import (
	"bytes"
	"net"
	"testing"
)

var magicPacket = []byte{
	255, 255, 255, 255, 255, 255,
	101, 172, 129, 19, 141, 63,
	101, 172, 129, 19, 141, 63,
	101, 172, 129, 19, 141, 63,
	101, 172, 129, 19, 141, 63,
	101, 172, 129, 19, 141, 63,
	101, 172, 129, 19, 141, 63,
	101, 172, 129, 19, 141, 63,
	101, 172, 129, 19, 141, 63,
	101, 172, 129, 19, 141, 63,
	101, 172, 129, 19, 141, 63,
	101, 172, 129, 19, 141, 63,
	101, 172, 129, 19, 141, 63,
	101, 172, 129, 19, 141, 63,
	101, 172, 129, 19, 141, 63,
	101, 172, 129, 19, 141, 63,
	101, 172, 129, 19, 141, 63,
}

func TestNewMagicPacket(t *testing.T) {
	hwAddr, err := net.ParseMAC("65:ac:81:13:8d:3f")
	if err != nil {
		t.Fatal(err)
	}
	got := NewMagicPacket(hwAddr)
	if !bytes.Equal(got, magicPacket) {
		t.Errorf("want %v, got %v", magicPacket, got)
	}
}
