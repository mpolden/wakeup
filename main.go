package main

import (
	"log"
	"net"
	"net/http"
	"os"

	flags "github.com/jessevdk/go-flags"
	"github.com/mpolden/wakeonlan/api"
)

func main() {
	var opts struct {
		CacheFile string `short:"c" long:"cache" description:"Path to cache file" required:"true" value-name:"FILE"`
		SourceIP  string `short:"b" long:"bind" description:"IP address to bind to when sending WOL packets" value-name:"IP"`
		Listen    string `short:"l" long:"listen" description:"Listen address" value-name:"ADDR" default:":8080"`
		StaticDir string `short:"s" long:"static" description:"Path to directory containing static assets" value-name:"DIR"`
	}
	_, err := flags.ParseArgs(&opts, os.Args)
	if err != nil {
		os.Exit(1)
	}

	sourceIP := net.ParseIP(opts.SourceIP)
	if sourceIP == nil {
		log.Fatalf("invalid ip: %s", opts.SourceIP)
	}

	api := api.New(opts.CacheFile)
	api.StaticDir = opts.StaticDir
	api.SourceIP = sourceIP
	log.Printf("Listening on %s", opts.Listen)
	if err := http.ListenAndServe(opts.Listen, api.Handler()); err != nil {
		log.Fatal(err)
	}
}
