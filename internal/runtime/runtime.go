package runtime

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/chuxorg/yanzi/internal/api/handlers"
	"github.com/chuxorg/yanzi/internal/api/models"
	apiserver "github.com/chuxorg/yanzi/internal/api/server"
)

const defaultRuntimeMode = "shared"
const defaultRuntimeAddr = "127.0.0.1:8080"

// Options captures the minimal runtime bootstrap inputs.
type Options struct {
	Addr            string
	Version         string
	RuntimeMode     string
	ShutdownTimeout time.Duration
	Dependencies    handlers.Dependencies
}

// Runtime owns a lightweight shared operational API server.
type Runtime struct {
	server          *apiserver.Server
	startedAt       time.Time
	runtimeMode     string
	shutdownTimeout time.Duration
}

// Instance represents a started runtime server.
type Instance struct {
	runtime   *Runtime
	listener  net.Listener
	errCh     chan error
	startedAt time.Time
}

// New constructs a runtime bootstrap wrapper around the operational API server.
func New(opts Options) *Runtime {
	addr := strings.TrimSpace(opts.Addr)
	if addr == "" {
		addr = defaultRuntimeAddr
	}
	runtimeMode := strings.TrimSpace(opts.RuntimeMode)
	if runtimeMode == "" {
		runtimeMode = defaultRuntimeMode
	}
	shutdownTimeout := opts.ShutdownTimeout
	if shutdownTimeout <= 0 {
		shutdownTimeout = 5 * time.Second
	}

	runtime := &Runtime{
		runtimeMode:     runtimeMode,
		shutdownTimeout: shutdownTimeout,
	}
	deps := opts.Dependencies
	deps.Version = opts.Version
	deps.RuntimeStatus = runtime.runtimeHealth

	runtime.server = apiserver.NewLocal(apiserver.LocalOptions{
		Addr:         addr,
		Version:      opts.Version,
		Dependencies: deps,
	})
	return runtime
}

// Start binds the runtime listener and begins serving requests in the background.
func (r *Runtime) Start() (*Instance, error) {
	if r == nil || r.server == nil || r.server.HTTPServer() == nil {
		return nil, errors.New("runtime is not initialized")
	}

	listener, err := net.Listen("tcp", r.server.HTTPServer().Addr)
	if err != nil {
		return nil, fmt.Errorf("start runtime listener: %w", err)
	}

	r.startedAt = time.Now().UTC()
	inst := &Instance{
		runtime:   r,
		listener:  listener,
		errCh:     make(chan error, 1),
		startedAt: r.startedAt,
	}

	go func() {
		err := r.server.Serve(listener)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			inst.errCh <- err
			return
		}
		inst.errCh <- nil
	}()

	return inst, nil
}

// Addr returns the bound listener address.
func (i *Instance) Addr() string {
	if i == nil || i.listener == nil {
		return ""
	}
	return i.listener.Addr().String()
}

// StartedAt returns the runtime startup timestamp.
func (i *Instance) StartedAt() time.Time {
	if i == nil {
		return time.Time{}
	}
	return i.startedAt
}

// Shutdown stops the runtime server with a bounded grace period.
func (i *Instance) Shutdown(ctx context.Context) error {
	if i == nil || i.runtime == nil || i.runtime.server == nil {
		return nil
	}
	shutdownCtx, cancel := context.WithTimeout(ctx, i.runtime.shutdownTimeout)
	defer cancel()
	err := i.runtime.server.Shutdown(shutdownCtx)
	if err == nil {
		i.runtime.startedAt = time.Time{}
	}
	return err
}

// Wait returns the background server error or nil after shutdown.
func (i *Instance) Wait() error {
	if i == nil || i.errCh == nil {
		return nil
	}
	return <-i.errCh
}

func (r *Runtime) runtimeHealth() *models.RuntimeHealth {
	if r == nil || r.startedAt.IsZero() {
		return nil
	}
	return &models.RuntimeHealth{
		Mode:      r.runtimeMode,
		StartedAt: r.startedAt.Format(time.RFC3339Nano),
	}
}
