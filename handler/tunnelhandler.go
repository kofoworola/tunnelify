package handler

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/kofoworola/tunnelify/config"
)

const bufferSize = 4096

var wg sync.WaitGroup

type TunnelHandler struct {
	incoming io.ReadWriter
	outgoing net.Conn

	serverURL   string
	httpVersion string

	closeConn CloseFunc

	cfg *config.Config
}

func NewTunnelHandler(cfg *config.Config, incoming io.ReadWriter, server string, httpVersion string, closeConn CloseFunc) *TunnelHandler {
	return &TunnelHandler{
		incoming:    incoming,
		serverURL:   server,
		httpVersion: httpVersion,
		closeConn:   closeConn,
		cfg:         cfg,
	}
}

func (h *TunnelHandler) Handle() {
	// get first req
	req, err := http.ReadRequest(bufio.NewReader(h.incoming))
	if err != nil {
		fmt.Printf("error reading request from connection: %v", err)
		return
	}
	// check the authorization
	if !checkAuthorization(h.cfg, req) {
		if err := writeResponse(
			h.incoming,
			req.Proto,
			"407 Proxy Authentication Required",
			http.Header{
				proxyAuthenticate: {`Basic realm="Access to the internal site"`},
			}); err != nil {
			fmt.Printf("error writing response: %v", err)
		}
		return

	}

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
