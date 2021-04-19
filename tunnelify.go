package main

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/kofoworola/tunnelify/config"
	"github.com/kofoworola/tunnelify/handler"
	"github.com/kofoworola/tunnelify/logging"
)

// TODO allow request on few IP
// TODO proxy authentication

type Server struct {
	config       *config.Config
	listener     net.Listener
	logger       *logging.Logger
	cleanupFuncs []func() error
}

func NewServer(cfg *config.Config) (*Server, error) {
	logger, err := logging.NewLogger(cfg)
	if err != nil {
		return nil, fmt.Errorf("error creating logger: %w", err)
	}

	listener, err := net.Listen("tcp", cfg.HostName)
	if err != nil {
		return nil, fmt.Errorf("error creating listener for %s: %w", cfg.HostName, err)
	}

	return &Server{
		listener:     listener,
		config:       cfg,
		logger:       logger,
		cleanupFuncs: []func() error{listener.Close},
	}, nil
}

func (p *Server) Start(ctx context.Context) error {
	// listen to new connections
listenLoop:
	for {
		select {
		case <-ctx.Done():
			break listenLoop
		default:
			c, err := p.listener.Accept()
			var h handler.ConnectionHandler
			if err != nil {
				p.logger.LogError("error accepting a new connection", err)
				break
			}
			// read first line of the connection and use an appropriate handler
			r := bufio.NewReader(c)
			reqLine, err := r.ReadBytes('\n')
			if err != nil {
				p.logger.LogError("error reading request line from connection", err)
				continue
			}

			// check the reqline for the handler to use
			reqDetails := strings.Split(string(reqLine), " ")
			if len(reqDetails) != 3 {
				p.logger.LogError("invalid request start line", nil)
				continue
			}

			logger := p.logger.With("action", reqDetails[0])
			cr := NewConnectionReader(reqLine, r, c)
			if reqDetails[0] == "CONNECT" {
				h = handler.NewTunnelHandler(p.config, cr, reqDetails[1], strings.TrimSpace(reqDetails[2]), c.Close)
			} else if reqDetails[0] != "CONNECT" && !strings.HasPrefix("/", reqDetails[1]) {
				h = handler.NewProxyHandler(cr, c.RemoteAddr().String(), p.config, c.Close)
			}
			go h.Handle(logger)
		}
	}
	return nil
}

func (p *Server) addToCleanups(cleanupFunc func() error) {
	p.cleanupFuncs = append(p.cleanupFuncs, cleanupFunc)
}
