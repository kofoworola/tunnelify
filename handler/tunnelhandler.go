package handler

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

const bufferSize = 4096

var wg sync.WaitGroup

type TunnelHandler struct {
	incoming io.ReadWriter
	outgoing net.Conn

	serverURL   string
	httpVersion string

	closeConn CloseFunc
}

func NewTunnelHandler(incoming io.ReadWriter, server string, httpVersion string, closeConn CloseFunc) (*TunnelHandler, error) {
	return &TunnelHandler{
		incoming:    incoming,
		serverURL:   server,
		httpVersion: httpVersion,
		closeConn:   closeConn,
	}, nil
}

func (h *TunnelHandler) Handle() {
	if h.outgoing == nil {
		c, err := h.setupOutbound()
		if err != nil {
			fmt.Printf("error setting up outbound connection %v\n", err)
		}
		defer c()
		response := fmt.Sprintf("%s 200 OK\n\n", h.httpVersion)
		h.incoming.Write([]byte(response))
	}
	wg.Add(2)
	go readAndWrite(h.incoming, h.outgoing)
	go readAndWrite(h.outgoing, h.incoming)

	// handle this properly because it is going to be impossible to close the connections
	// from this (the tunnel) end atm
	wg.Wait()
	h.closeConn()
}

func readAndWrite(readFrom io.Reader, writeTo io.Writer) {
	defer wg.Done()
	for {
		var shouldBreak bool
		dat := make([]byte, bufferSize)
		n, err := readFrom.Read(dat)
		if err != nil {
			shouldBreak = true
			if err != io.EOF {
				fmt.Printf("error reading data: %v\n", err)
				break
			}
		}
		dat = dat[:n]

		if _, err := writeTo.Write(dat); err != nil {
			fmt.Printf("error writing: %v\n", err)
			break
		}
		if shouldBreak {
			break
		}
	}
}

func (h *TunnelHandler) setupOutbound() (func() error, error) {
	conn, err := net.DialTimeout("tcp", h.serverURL, time.Second*30)
	if err != nil {
		return nil, err
	}
	h.outgoing = conn
	return conn.Close, err
}
