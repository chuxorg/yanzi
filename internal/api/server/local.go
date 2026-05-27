package server

import (
	"time"

	"github.com/chuxorg/yanzi/internal/api/handlers"
	"github.com/chuxorg/yanzi/internal/api/routes"
)

// LocalOptions captures the local-only operational API server construction inputs.
type LocalOptions struct {
	Addr              string
	Version           string
	ReadHeaderTimeout time.Duration
	Dependencies      handlers.Dependencies
}

// NewLocal constructs a server wired to the current operational API route foundation.
func NewLocal(opts LocalOptions) *Server {
	deps := opts.Dependencies
	deps.Version = opts.Version
	return New(Options{
		Addr:              opts.Addr,
		Handler:           routes.NewHandler(deps),
		ReadHeaderTimeout: opts.ReadHeaderTimeout,
	})
}
