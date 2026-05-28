package server

import (
	"context"
	"net"
	"net/http"
	"time"
)

const defaultReadHeaderTimeout = 5 * time.Second

// Options captures the minimal HTTP server configuration for the operational API.
type Options struct {
	Addr              string
	Handler           http.Handler
	ReadHeaderTimeout time.Duration
}

// Server is the lightweight internal HTTP server foundation for the operational API.
type Server struct {
	handler    http.Handler
	httpServer *http.Server
}

// New constructs an internal operational API server with conservative defaults.
func New(opts Options) *Server {
	opts = opts.withDefaults()
	return &Server{
		handler: opts.Handler,
		httpServer: &http.Server{
			Addr:              opts.Addr,
			Handler:           opts.Handler,
			ReadHeaderTimeout: opts.ReadHeaderTimeout,
		},
	}
}

// Handler returns the server handler used to process HTTP requests.
func (s *Server) Handler() http.Handler {
	if s == nil {
		return nil
	}
	return s.handler
}

// HTTPServer exposes the underlying net/http server for controlled internal use.
func (s *Server) HTTPServer() *http.Server {
	if s == nil {
		return nil
	}
	return s.httpServer
}

// Serve starts the HTTP server on the provided listener.
func (s *Server) Serve(listener net.Listener) error {
	if s == nil || s.httpServer == nil {
		return nil
	}
	return s.httpServer.Serve(listener)
}

// Shutdown gracefully stops the HTTP server.
func (s *Server) Shutdown(ctx context.Context) error {
	if s == nil || s.httpServer == nil {
		return nil
	}
	return s.httpServer.Shutdown(ctx)
}

func (o Options) withDefaults() Options {
	if o.Handler == nil {
		o.Handler = http.NotFoundHandler()
	}
	if o.ReadHeaderTimeout <= 0 {
		o.ReadHeaderTimeout = defaultReadHeaderTimeout
	}
	return o
}
