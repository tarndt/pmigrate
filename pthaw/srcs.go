package main

import (
	"io"
	"net"
	"os"
	"strings"

	"lib/errs"
)

func getSourceReader(src string) (io.ReadCloser, error) {
	if src == "stdin" || src == "" { //Stdin
		return os.Stdin, nil
	} else if strings.ContainsRune(src, ':') { //Network and unix sockets
		args := strings.Split(src, ":")
		if len(args) < 2 {
			return nil, errs.New("Network/IPC destinations must be in the form: proto:arg1:argN...")
		}
		proto := strings.ToLower(strings.TrimSpace(args[0]))
		var addr string
		switch proto {
		case "tcp", "udp":
			if len(args) != 2 {
				return nil, errs.New("Network destinations must be in the form: tcp|udp:port.")
			}
			addr = ":" + strings.TrimSpace(args[1])
		case "unix":
			if len(args) != 2 {
				return nil, errs.New("IPC (Unix socket) destinations must be in the form: unix:socketpath.")
			}
			addr = strings.TrimSpace(args[1])
		default:
			return nil, errs.New("Unknown network/ICP destination protcol: %q. Use: tcp,udp or unix.")
		}

		//Wait for/build connection
		var conn net.Conn
		switch proto {
		case "udp": //Not connection oriented
			UDPAddr, err := net.ResolveUDPAddr("udp", addr)
			if err != nil {
				return nil, errs.Append(err, "Failed to resolve UDP address: %s", addr)
			}
			var UDPConn *net.UDPConn
			if UDPConn, err = net.ListenUDP("udp", UDPAddr); err != nil {
				return nil, errs.Append(err, "Failed to listen for UDP packets")
			}
			UDPConn.SetReadBuffer(8 * 1024 * 1024)
			conn = UDPConn
		default: //Connection oriented protcols need to wait for client
			listener, err := net.Listen(proto, addr)
			if err != nil {
				return nil, errs.Append(err, "Listen: %s/%s failed", proto, addr)
			}
			defer listener.Close()
			if conn, err = listener.Accept(); err != nil {
				return nil, errs.Append(err, "Accept: %s/%s failed", proto, addr)
			}
		}
		return conn, nil
	}
	//File
	return os.Open(src)
}
