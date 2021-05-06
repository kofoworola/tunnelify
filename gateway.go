package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"

	"github.com/kofoworola/tunnelify/config"
	"github.com/kofoworola/tunnelify/handler"
	"github.com/kofoworola/tunnelify/logging"
)

const defaultBufferSize = 2048

type listenerGateway struct {
	net.Listener

	connChan    chan net.Conn
	listenerErr error

	config *config.Config
	logger *logging.Logger
}

func NewGateway(cfg *config.Config) (*listenerGateway, error) {
	logger, err := logging.NewLogger(cfg)
	if err != nil {
		return nil, fmt.Errorf("error creating logger: %w", err)
	}

	listener, err := net.Listen("tcp", ":"+cfg.Port)
	if err != nil {
		return nil, fmt.Errorf("error creating listener for %s: %w", cfg.Port, err)
	}

	return &listenerGateway{
		Listener: listener,
		connChan: make(chan net.Conn, 1),
		config:   cfg,
		logger:   logger,
	}, nil
}

func (l *listenerGateway) Accept() (net.Conn, error) {
	if l.listenerErr != nil {
		return nil, l.listenerErr
	}
	c := <-l.connChan
	return c, nil
}

// Start listnes for new connections from the core listener,
// then determines how to handle it. Either passing it to the proxy handler,
// the tunnel handler, or sending it to the connection channel for it's own Accept() method.
func (l *listenerGateway) Start() error {
	// listen to new connections
	for {
		c, err := l.Listener.Accept()
		if err != nil {
			l.listenerErr = err
			l.logger.LogError("error accepting a new connection", err)
			break
		}
		var h handler.ConnectionHandler

		// read first line of the connection and use an appropriate handler
		r := bufio.NewReaderSize(c, defaultBufferSize)
		reqLine, err := r.ReadBytes('\n')
		if err != nil {
			l.logger.LogError("error reading request line from connection", err)
			continue
		}

		// use the length of the first line to determine the content of the buffer
		// and fetch that to prepend to the connection
		bufferContent := make([]byte, defaultBufferSize-len(reqLine))
		n, err := r.Read(bufferContent)
		if err != nil {
			l.logger.LogError("error reading request from connection", err)
			continue
		}

		// make sure the length of what was read is the same as the length of the bufferContent
		// if not trim it
		if n < len(bufferContent) {
			bufferContent = bufferContent[:n]
		}

		// check the reqline for the handler to use
		reqDetails := strings.Split(string(reqLine), " ")
		if len(reqDetails) != 3 {
			l.logger.LogError("invalid request start line", nil)
			continue
		}

		logger := l.logger.With("action", reqDetails[0])
		cr := NewConnectionReader(c, reqLine, bufferContent)
		if reqDetails[0] == "CONNECT" {
			h = handler.NewTunnelHandler(l.config, cr, reqDetails[1], strings.TrimSpace(reqDetails[2]), c.Close)
		} else if reqDetails[0] != "CONNECT" && !strings.HasPrefix(reqDetails[1], "/") {
			h = handler.NewProxyHandler(cr, c.RemoteAddr().String(), l.config, c.Close)
		} else {
			l.connChan <- cr
		}

		if h != nil {
			// check if allowed
			if !l.config.ShouldAllowIP(c.RemoteAddr().String()) {
				handler.WriteResponse(c, "HTTP/1.1", "403 Forbidden", nil)
				c.Close()
				continue
			}
			go h.Handle(logger)
		}
	}
	return nil
}
