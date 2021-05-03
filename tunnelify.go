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

// TODO allow request on few IP

type Server struct {
	config   *config.Config
	listener net.Listener
	logger   *logging.Logger
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
		listener: listener,
		config:   cfg,
		logger:   logger,
	}, nil
}

func (s *Server) Shutdown() {
	s.listener.Close()
}

func (s *Server) Start() error {
	// listen to new connections
	for {
		c, err := s.listener.Accept()
		if err != nil {
			s.logger.LogError("error accepting a new connection", err)
			break
		}
		var h handler.ConnectionHandler

		// check if allowed
		if !s.config.ShouldAllowIP(c.RemoteAddr().String()) {
			c.Close()
			continue
		}
		// read first line of the connection and use an appropriate handler
		r := bufio.NewReader(c)
		reqLine, err := r.ReadBytes('\n')
		if err != nil {
			s.logger.LogError("error reading request line from connection", err)
			continue
		}

		// check the reqline for the handler to use
		reqDetails := strings.Split(string(reqLine), " ")
		if len(reqDetails) != 3 {
			s.logger.LogError("invalid request start line", nil)
			continue
		}

		logger := s.logger.With("action", reqDetails[0])
		cr := NewConnectionReader(reqLine, r, c)
		if reqDetails[0] == "CONNECT" {
			h = handler.NewTunnelHandler(s.config, cr, reqDetails[1], strings.TrimSpace(reqDetails[2]), c.Close)
		} else if reqDetails[0] != "CONNECT" && !strings.HasPrefix("/", reqDetails[1]) {
			h = handler.NewProxyHandler(cr, c.RemoteAddr().String(), s.config, c.Close)
		}
		go h.Handle(logger)
	}
	return nil
}
