package handler

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

var (
	invalidRequestFormat = errors.New("invalid request format")
)

const (
	proxyConnectionKey = "Proxy-Connection"
)

type connWrapper struct {
	net.Conn
	mu sync.Mutex
}

type Request struct {
	URI     string
	Version string
	Method  string
	Headers map[string]string
	Body    io.Reader
}

type ProxyHandler struct {
	incoming io.ReadWriter
	outgoing *connWrapper

	connClose CloseFunc
}

//TODO add timeout to config
// TODO add dial timout to config
// TODO add check redirect to config
func NewProxyHandler(reader io.ReadWriter, serverURL string, closeFunc CloseFunc) *ProxyHandler {
	// the reason we don't dial initially to the server is to prevent a bottleneck
	// for multiple proxy connections coming in
	return &ProxyHandler{
		incoming:  reader,
		connClose: closeFunc,
	}
}

// TODO handle errors properly in goroutine using a context passed logger or something
// TODO handle
func (p *ProxyHandler) Handle() {
	for {
		req, err := http.ReadRequest(bufio.NewReader(p.incoming))
		if err != nil {
			// if it is an EOF error, close the connection and carry on
			if err == io.EOF {
				p.connClose()
				return
			}
			fmt.Printf("error parsing request: %v", err)
			return
		}
		// setup outgoing connection if it hasn't been setUp
		if p.outgoing == nil {
			addr := fmt.Sprintf("%s:%s", req.URL.Host, req.URL.Scheme)
			conn, err := net.DialTimeout("tcp", addr, time.Second*30)
			if err != nil {
				fmt.Printf("error dialing to server: %v", err)
				p.connClose()
				break
			}
			p.outgoing = &connWrapper{Conn: conn}
			go p.listenToServerIncoming()
			defer conn.Close()
		}

		if err := p.prepareRequest(req); err != nil {
			fmt.Printf("error preparing request: %v", err)
			return
		}
		p.outgoing.mu.Lock()
		if err := req.Write(p.outgoing); err != nil {
			fmt.Printf("error writing to server: %v", err)
			break
		}
		p.outgoing.mu.Unlock()
		shouldClose := p.shouldCloseConnection(req)
		if shouldClose {
			break
		}
	}
	if err := p.connClose(); err != nil {
		fmt.Printf("error closing connection: %v", err)
	}
}

func (p *ProxyHandler) listenToServerIncoming() {
	reader := bufio.NewReader(p.outgoing)
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Printf("error reading response: %v\n", err)
			break
		}
		// TODO fix code reaching here
		if _, err := p.incoming.Write(line); err != nil {
			fmt.Printf("error writing response: %v\n", err)
		}
	}
}

// prepareRequest prepares the request to be sent to the server
// by removing the RequestURI and setting the req.URL
// then formating the headers
// TODO handle headers and remove all proxy based headers
func (p *ProxyHandler) prepareRequest(req *http.Request) error {
	url, err := url.Parse(req.RequestURI)
	if err != nil {
		return err
	}
	req.URL = url
	req.RequestURI = ""
	delete(req.Header, "Proxy-Connection")
	return nil
}

// TODO drain the response
func (p *ProxyHandler) writeResponse(resp *http.Response) {
	defer resp.Body.Close()
	// write the status line e.g HTTP/1.1 404 Not Found
	p.incoming.Write([]byte(fmt.Sprintf("%s %s\n", resp.Proto, resp.Status)))

	// write the headers
	for key, val := range resp.Header {
		headerString := fmt.Sprintf("%s:%s\n", key, strings.Join(val, ","))
		p.incoming.Write([]byte(headerString))
	}
	p.incoming.Write([]byte("\n"))
	io.Copy(p.incoming, resp.Body)
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
	if found {
		return true
	}
	return false
}
