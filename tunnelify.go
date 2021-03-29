package tunnelify

import (
	"context"
	"fmt"
	"net"

	"github.com/kofoworola/tunnelify/config"
	"go.uber.org/zap"
)

type Proxy struct {
	config       *config.Config
	listener     net.Listener
	logger       *zap.Logger
	cleanupFuncs []func() error
}

func NewProxy(cfg *config.Config) (*Proxy, error) {
	logger, err := zap.NewProduction(zap.WithCaller(true))
	if err != nil {
		return nil, fmt.Errorf("error creating logger: %w", err)
	}

	listener, err := net.Listen("tcp", cfg.HostName)
	if err != nil {
		return nil, fmt.Errorf("error creating listener for %s: %w", cfg.HostName, err)
	}

	return &Proxy{
		listener:     listener,
		config:       cfg,
		logger:       logger,
		cleanupFuncs: []func() error{listener.Close},
	}, nil

}

func (p *Proxy) Start(ctx context.Context) error {
	// listen to new connections
listenLoop:
	for {
		select {
		case <-ctx.Done():
			break listenLoop
		default:
			conn, err := p.listener.Accept()
			if err != nil {
				p.logger.Error("error accepting a new connection")
			}
			go p.handleConnection(conn)
		}
	}
	return nil
}

func (p *Proxy) handleConnection(conn net.Conn) error {
	return nil
}

func (p *Proxy) addToCleanups(cleanupFunc func() error) {
	p.cleanupFuncs = append(p.cleanupFuncs, cleanupFunc)
}
