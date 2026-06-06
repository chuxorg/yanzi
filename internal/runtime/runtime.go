package runtime

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/chuxorg/yanzi/internal/api/handlers"
	"github.com/chuxorg/yanzi/internal/api/models"
	apiserver "github.com/chuxorg/yanzi/internal/api/server"
	"github.com/chuxorg/yanzi/internal/auth"
	"github.com/chuxorg/yanzi/internal/config"
	"github.com/chuxorg/yanzi/internal/packs"
	"github.com/chuxorg/yanzi/internal/storage"
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
	// Provider is an optional pre-initialized storage provider. When set,
	// it is used directly instead of opening a new provider per request.
	Provider      storage.Provider
	APIKeyStore   auth.APIKeyStore
	AuthConfig    config.AuthConfig
	OIDCValidator *auth.OIDCValidator
	// TLSCert and TLSKey are paths to PEM files for TLS. Both must be set to
	// enable HTTPS; if both are empty the server runs plain HTTP.
	TLSCert    string
	TLSKey     string
	PackStore  packs.PackStore
}

// Runtime owns a lightweight shared operational API server.
type Runtime struct {
	server          *apiserver.Server
	startedAt       time.Time
	runtimeMode     string
	shutdownTimeout time.Duration
	tlsCert         string
	tlsKey          string
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
		tlsCert:         opts.TLSCert,
		tlsKey:          opts.TLSKey,
	}
	deps := opts.Dependencies
	deps.Version = opts.Version
	deps.RuntimeStatus = runtime.runtimeHealth

	// Wire the pre-initialized provider so handlers return it directly
	// instead of opening a new connection per request.
	if opts.Provider != nil {
		p := opts.Provider
		deps.OpenProvider = func(_ context.Context, _ config.Config) (storage.Provider, error) {
			return p, nil
		}
	}

	deps.APIKeyStore = opts.APIKeyStore
	deps.AuthConfig = opts.AuthConfig
	deps.OIDCValidator = opts.OIDCValidator
	deps.PackStore = opts.PackStore

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

	tcpListener, err := net.Listen("tcp", r.server.HTTPServer().Addr)
	if err != nil {
		return nil, fmt.Errorf("start runtime listener: %w", err)
	}

	var listener net.Listener = tcpListener
	if r.tlsCert != "" && r.tlsKey != "" {
		cert, err := tls.LoadX509KeyPair(r.tlsCert, r.tlsKey)
		if err != nil {
			_ = tcpListener.Close()
			return nil, fmt.Errorf("load TLS certificate: %w", err)
		}
		tlsCfg := &tls.Config{
			Certificates: []tls.Certificate{cert},
			MinVersion:   tls.VersionTLS12,
		}
		listener = tls.NewListener(tcpListener, tlsCfg)
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
