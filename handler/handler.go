package handler

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/kofoworola/tunnelify/config"
	"github.com/kofoworola/tunnelify/logging"
)

type CloseFunc func() error

type ConnectionHandler interface {
	Handle(logger *logging.Logger)
}

func WriteResponse(out io.Writer, version string, status string, header http.Header) error {
	var builder strings.Builder
	if _, err := builder.WriteString(fmt.Sprintf("%s %s\n", version, status)); err != nil {
		return err
	}

	if _, err := builder.WriteString(fmt.Sprintf("%s: 0\n", contentLength)); err != nil {
		return err
	}
	for key, val := range header {
		for _, item := range val {
			headerLine := fmt.Sprintf("%s: %s\n", key, item)
			if _, err := builder.WriteString(headerLine); err != nil {
				return err
			}
		}
	}
	builder.WriteString("\n")
	if _, err := out.Write([]byte(builder.String())); err != nil {
		return err
	}
	return nil
}

// checkAuthorization checks if the authorization string matches the request
func checkAuthorization(cfg *config.Config, req *http.Request) bool {
	if cfg.HasAuth() {
		authHeader, ok := req.Header[proxyAuthorization]
		if !ok {
			return false
		}

		authString := authHeader[0]
		return cfg.CheckAuthString(authString)
	}
	return true
}
