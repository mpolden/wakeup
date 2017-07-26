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

func TestBridgeRead(t *testing.T) {
	b := Bridge{conn: bytes.NewReader(magicPacket)}
	mp, err := b.ReadMagicPacket()
	if err != nil {
		t.Fatal(err)
	}
	want := "65:ac:81:13:8d:3f"
	if got := mp.HardwareAddr().String(); got != want {
		t.Errorf("want %s, got %s", want, got)
	}

	b.conn = bytes.NewReader([]byte{1, 2, 3})
	if _, err := b.ReadMagicPacket(); err == nil {
		t.Fatal("expected error")
	}
}

func TestBridgeForward(t *testing.T) {
	var target net.HardwareAddr
	wake := func(src net.IP, hwAddr net.HardwareAddr) error {
		target = hwAddr
		return nil
	}
	b := Bridge{
		conn:     bytes.NewReader(magicPacket),
		wakeFunc: wake,
	}
	if _, err := b.Forward(nil); err != nil {
		t.Fatal(err)
	}
	want := "65:ac:81:13:8d:3f"
	if got := target.String(); got != want {
		t.Errorf("want %s, got %s", want, got)
	}
}

func TestBridgeForwardPreventsLoop(t *testing.T) {
	n := 0
	wake := func(src net.IP, hwAddr net.HardwareAddr) error {
		n += 1
		return nil
	}
	var buf bytes.Buffer
	b := Bridge{
		conn:     &buf,
		wakeFunc: wake,
	}
	// Same magic packet is received a second time, this likely means that we sent it ourself
	for i := 0; i < 2; i++ {
		buf.Write(magicPacket)
		if _, err := b.Forward(nil); err != nil {
			t.Fatal(err)
		}
	}
	if n != 1 {
		t.Errorf("want 1 wake up, got %d", n)
	}
}
