package tunnelify

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/kofoworola/tunnelify/config"
	"github.com/kofoworola/tunnelify/handler"
	"go.uber.org/zap"
)

// TODO work on logging
// TODO allow request on few IP

type Server struct {
	config       *config.Config
	listener     net.Listener
	logger       *zap.Logger
	cleanupFuncs []func() error
}

func NewServer(cfg *config.Config) (*Server, error) {
	logger, err := zap.NewProduction(zap.WithCaller(true))
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
			rw, err := p.listener.Accept()
			var h handler.ConnectionHandler
			if err != nil {
				p.logger.Error("error accepting a new connection")
			}
			// read first line of the connection and use an appropriate handler
			bufReader := bufio.NewReader(rw)
			reqLine, err := bufReader.ReadBytes('\n')
			if err != nil {
				p.logger.Error(fmt.Sprintf("error reading request line from connection: %v", err))

			}
			// check the reqline for the handler to use
			reqDetails := strings.Split(string(reqLine), " ")
			if len(reqDetails) != 3 {
				p.logger.Error("invalid request start line")
			}

			c := NewReadWrapper(bufReader)
			c.Write(reqLine)
			if reqDetails[0] == "CONNECT" {
				h = nil
				continue
				// start tunnel
			} else if reqDetails[0] != "CONNECT" && !strings.HasPrefix("/", reqDetails[1]) {
				// most likely http/1 so use the proxy handler
				h = handler.NewProxyHandler(rw)
			}
			go h.Handle()
		}
	}
	return nil
}

func (p *Server) addToCleanups(cleanupFunc func() error) {
	p.cleanupFuncs = append(p.cleanupFuncs, cleanupFunc)
}
