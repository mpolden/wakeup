# wakeup

[![Build Status](https://travis-ci.org/mpolden/wakeup.svg?branch=master)](https://travis-ci.org/mpolden/wakeup)

`wakeup` provides a small HTTP API and JavaScript front-end for
sending [Wake-on-LAN](https://en.wikipedia.org/wiki/Wake-on-LAN) messages to a
target device.

## Usage

```
$ wakeup -h
Usage:
  wakeup [OPTIONS]

Application Options:
  -c, --cache=FILE     Path to cache file
  -b, --bind=IP        IP address to bind to when sending WOL packets
  -l, --listen=ADDR    Listen address (default: :8080)
  -s, --static=DIR     Path to directory containing static assets

Help Options:
  -h, --help           Show this help message
```

## API

Wake a device:

`$ curl -XPOST -d '{"macAddress":"AB:CD:EF:12:34:56"}' http://localhost:8080/api/v1/wake`

A name for the device can also be provided, to make it easy to identify later:

`$ curl -XPOST -d '{"name":"foo","macAddress":"AB:CD:EF:12:34:56"}' http://localhost:8080/api/v1/wake`

List devices that have previously been woken:

```
$ curl -s http://localhost:8080/api/v1/wake | jq .
{
  "devices": [
    {
      "name": "foo",
      "macAddress": "AB:CD:EF:12:34:56"
    }
  ]
}
```

Delete a device:

`$ curl -XDELETE -d '{"macAddress":"AB:CD:EF:12:34:56"}' http://localhost:8080/api/v1/wake`

## Front-end

A basic JavaScript front-end is included. It can be served by `wakeup` by
passing the path to `static` as the `-s` option.

![Front-end screenshot](static/screenshot.png)
