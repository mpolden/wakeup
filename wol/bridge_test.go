package wol

import (
	"bytes"
	"io"
	"net"
	"testing"
)

type mockConn struct{ io.Reader }

func (c *mockConn) Close() error { return nil }

func TestBridgeRead(t *testing.T) {
	b := Bridge{conn: &mockConn{bytes.NewReader(magicPacket)}}
	mp, err := b.read()
	if err != nil {
		t.Fatal(err)
	}
	want := "65:ac:81:13:8d:3f"
	if got := mp.HardwareAddr().String(); got != want {
		t.Errorf("want %s, got %s", want, got)
	}

	b.conn = &mockConn{bytes.NewReader([]byte{1, 2, 3})}
	want = "invalid magic packet: 010203"
	if _, err := b.read(); err.Error() != want {
		t.Errorf("got %q, want %q", err.Error(), want)
	}
}

func TestBridgeForward(t *testing.T) {
	var target net.HardwareAddr
	wake := func(src net.IP, hwAddr net.HardwareAddr) error {
		target = hwAddr
		return nil
	}
	b := Bridge{
		conn:     &mockConn{bytes.NewReader(magicPacket)},
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
		conn:     &mockConn{&buf},
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
