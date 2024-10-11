package apiserver

import (
	"net/http"
	"time"

	"github.com/rs/xid"
)

type HTTPServer struct {
	ID string
	*http.Server
}

func NewServer(handler http.Handler, addr string) *HTTPServer {
	return &HTTPServer{
		xid.New().String(),
		&http.Server{
			Addr:           ":" + addr,
			ReadTimeout:    10 * time.Second,
			WriteTimeout:   10 * time.Second,
			MaxHeaderBytes: 1 << 20,
			Handler:        handler,
		},
	}
}
