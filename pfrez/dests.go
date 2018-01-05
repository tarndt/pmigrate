package main

import (
	"io"
	"net"
	"os"
	"strings"
	"time"

	"lib/errs"
)

type nopCloser struct {
	io.Writer
}

func (nopCloser) Close() error { return nil }

func newNopWriteCloser(wtr io.Writer) io.WriteCloser {
	return nopCloser{wtr}
}

func getDestWriter(dest string, dialTimeout time.Duration) (io.WriteCloser, error) {
	if dest == "stdout" || dest == "" { //Stdout
		return os.Stdout, nil
	} else if strings.ContainsRune(dest, ':') { //Network and unix sockets
		args := strings.Split(dest, ":")
		if len(args) < 2 {
			return nil, errs.New("Network/IPC destinations must be in the form: proto:arg1:argN...")
		}
		proto := strings.ToLower(strings.TrimSpace(args[0]))
		var addr string
		switch proto {
		case "tcp", "udp":
			if len(args) != 3 {
				return nil, errs.New("Network destinations must be in the form: tcp|udp:host:port.")
			}
			addr = strings.TrimSpace(args[1]) + ":" + strings.TrimSpace(args[2])
		case "unix":
			if len(args) != 2 {
				return nil, errs.New("IPC (Unix socket) destinations must be in the form: unix:socketpath.")
			}
			addr = strings.TrimSpace(args[1])
		default:
			return nil, errs.New("Unknown network/ICP destination protcol: %q. Use: tcp,udp or unix.")
		}
		if dialTimeout > 0 {
			return net.DialTimeout(proto, addr, dialTimeout)
		}
		return net.Dial(proto, addr)
	}
	//File
	return os.Create(dest)
}
