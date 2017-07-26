package main

import (
	"log"
	"net"
	"os"
	"strings"

	flags "github.com/jessevdk/go-flags"
	"github.com/mpolden/wakeup/wol"
)

func main() {
	var opts struct {
		ListenAddr  string `short:"l" long:"listen" description:"Listen address to use when listening for WOL packets" value-name:"IP" default:"0.0.0.0:9"`
		ForwardAddr string `short:"o" long:"forward" description:"Address of interface where received WOL packets should be forwarded" required:"true" value-name:"IP"`
	}
	_, err := flags.ParseArgs(&opts, os.Args)
	if err != nil {
		os.Exit(1)
	}

	forwardAddr := net.ParseIP(opts.ForwardAddr)
	if forwardAddr == nil {
		log.Fatalf("invalid ip: %s", opts.ForwardAddr)
	}

	b, err := wol.Listen(opts.ListenAddr)
	if err != nil {
		log.Fatal(err)
	}
	for {
		sent, err := b.Forward(forwardAddr)
		if err != nil {
			log.Fatal(err)
		}
		if sent != nil {
			log.Printf("Forwarded magic packet for %s to %s", strings.ToUpper(sent.HardwareAddr().String()), forwardAddr)
		}
	}
}
