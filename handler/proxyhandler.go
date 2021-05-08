package handler

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"

	"github.com/kofoworola/tunnelify/config"
	"github.com/kofoworola/tunnelify/logging"
)

// Header Keys
const (
	proxyConnectionKey = "Proxy-Connection"
	forwardedForKey    = "X-Forwarded-For"
	forwardedHost      = "X-Forwarded-Host"
	proxyAuthorization = "Proxy-Authorization"
	proxyAuthenticate  = "Proxy-Authenticate"
	contentLength      = "Content-Length"
)

type Request struct {
	URI     string
	Version string
	Method  string
	Headers map[string]string
	Body    io.Reader
}

type ProxyHandler struct {
	incoming io.ReadWriter
	outgoing net.Conn

	connClose CloseFunc

	originIP string
	cfg      *config.Config
}

func NewProxyHandler(reader io.ReadWriter, originIp string, config *config.Config, closeFunc CloseFunc) *ProxyHandler {
	// the reason we don't dial initially to the server is to prevent a bottleneck
	// for multiple proxy connections coming in
	return &ProxyHandler{
		incoming:  reader,
		connClose: closeFunc,
		originIP:  originIp,
		cfg:       config,
	}
}

func (p *ProxyHandler) Handle(logger *logging.Logger) {
	logger = logger.With("type", "proxy")

	for {
		req, err := http.ReadRequest(bufio.NewReader(p.incoming))
		if err != nil {
			// if it is an EOF error, close the connection and carry on
			if err == io.EOF {
				p.connClose()
				return
			}
			logger.Warn("error parsing request", nil)
			return
		}
		logger.Debug("received new request")

		// setup outgoing connection if it hasn't been setUp
		if p.outgoing == nil {
			addr := fmt.Sprintf("%s:%s", req.URL.Host, req.URL.Scheme)
			conn, err := net.DialTimeout("tcp", addr, p.cfg.Timeout)
			if err != nil {
				logger.Warn("error dialing destination server", nil)
				p.connClose()
				break
			}
			p.outgoing = conn
			go p.listenToServerIncoming(logger)
			defer conn.Close()
		}

		// check the authorization
		if !checkAuthorization(p.cfg, req) {
			logger.Debug("request not authorized")
			if err := WriteResponse(
				p.incoming,
				req.Proto,
				"407 Proxy Authentication Required",
				http.Header{
					proxyAuthenticate: {`Basic realm="Access to the internal site"`},
				}); err != nil {
				logger.Warn("error writing response", nil)
			}
			continue
		}

		if err := p.prepareRequest(req); err != nil {
			logger.Warn("error forwarding request", err)
			continue
		}
		if err := req.Write(p.outgoing); err != nil {
			logger.Warn("error forwarding request", err)
			continue
		}
		shouldClose := p.shouldCloseConnection(req)
		if shouldClose {
			break
		}
	}
	if err := p.connClose(); err != nil {
		logger.Warn("error closing connection", nil)
	}
}

func (p *ProxyHandler) listenToServerIncoming(logger *logging.Logger) {
	reader := bufio.NewReader(p.outgoing)
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			logger.Warn("could not read response from server", err)
			break
		}
		// TODO fix code reaching here
		if _, err := p.incoming.Write(line); err != nil {
			logger.Warn("error writing to client", err)
		}
	}
}

// prepareRequest prepares the request to be sent to the server
// by removing the RequestURI and setting the req.URL
// then formating the headers
func (p *ProxyHandler) prepareRequest(req *http.Request) error {
	url, err := url.Parse(req.RequestURI)
	if err != nil {
		return err
	}
	req.URL = url
	req.RequestURI = ""
	delete(req.Header, proxyConnectionKey)

	// add origin ip if enabled in config
	if !p.cfg.HideIP {
		forwarded, ok := req.Header[forwardedForKey]
		if !ok || len(forwarded) < 1 {
			req.Header.Set(forwardedForKey, p.originIP)
		} else {
			req.Header.Set(forwardedForKey, fmt.Sprintf("%s, %s", forwarded[0], p.originIP))
		}
		req.Header.Set(forwardedHost, req.Host)
	}
	return nil
}

func (p *ProxyHandler) shouldCloseConnection(req *http.Request) bool {
	val, ok := req.Header[proxyConnectionKey]
	// key doesn't exist in headers, don't close the connection
	if !ok {
		return false
	}
	found := false
	for _, i := range val {
		if i == "close" {
			found = true
			break
		}
	}
	return found
}
