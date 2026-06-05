package cmd

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/chuxorg/yanzi/internal/auth"
	"github.com/chuxorg/yanzi/internal/config"
	"github.com/chuxorg/yanzi/internal/runtime"
	"github.com/chuxorg/yanzi/internal/storage/registry"
)

func bootstrapAdminKey(ctx context.Context, store auth.APIKeyStore) error {
	keys, err := store.ListKeys(ctx)
	if err != nil {
		return fmt.Errorf("list keys: %w", err)
	}
	if len(keys) > 0 {
		return nil
	}

	_, fullKey, err := store.CreateKey(ctx, "bootstrap-admin", auth.ScopeAdmin, false)
	if err != nil {
		return fmt.Errorf("create bootstrap key: %w", err)
	}

	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════╗")
	fmt.Println("║           YANZI BOOTSTRAP API KEY                   ║")
	fmt.Println("║                                                      ║")
	fmt.Println("║  No API keys found. A bootstrap admin key has been  ║")
	fmt.Println("║  created. Copy it now — it will not be shown again. ║")
	fmt.Println("║                                                      ║")
	fmt.Printf( "║  Key:   %-44s ║\n", fullKey)
	fmt.Println("║  Scope: admin                                        ║")
	fmt.Println("║                                                      ║")
	fmt.Println("║  Use this key to create additional keys via:        ║")
	fmt.Println("║  POST /v0/keys                                      ║")
	fmt.Println("╚══════════════════════════════════════════════════════╝")
	fmt.Println()
	return nil
}

// RunServe starts the shared runtime server in the foreground.
func RunServe(args []string, version string) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	return runServe(ctx, args, version)
}

func runServe(ctx context.Context, args []string, version string) error {
	fs := flag.NewFlagSet("serve", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	host := fs.String("host", "127.0.0.1", "Host address to bind the HTTP server")
	addr := fs.String("addr", "127.0.0.1:8080", "listen address (host:port)")
	grace := fs.Duration("shutdown-timeout", 5*time.Second, "graceful shutdown timeout")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return errors.New("usage: yanzi serve [--host address] [--addr host:port] [--shutdown-timeout duration]")
	}

	_, port, err := net.SplitHostPort(strings.TrimSpace(*addr))
	if err != nil {
		return fmt.Errorf("invalid --addr %q: %w", *addr, err)
	}
	listenAddr := net.JoinHostPort(strings.TrimSpace(*host), port)

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	providerName := config.EffectiveStorageProvider(cfg)
	provider, err := registry.Open(ctx, cfg, registry.Options{})
	if err != nil {
		return err
	}
	defer func() { _ = provider.Close() }()

	fmt.Printf("storage provider: %s\n", providerName)

	var keyStore auth.APIKeyStore
	if ks, ok := provider.(auth.APIKeyStore); ok {
		keyStore = ks
	}

	if cfg.Auth.Enabled {
		fmt.Println("auth: enabled")
		if keyStore != nil {
			if err := bootstrapAdminKey(ctx, keyStore); err != nil {
				return fmt.Errorf("bootstrap admin key: %w", err)
			}
		}
	} else {
		fmt.Println("auth: disabled (all requests permitted)")
	}

	var oidcValidator *auth.OIDCValidator
	if cfg.Auth.OIDC.Enabled {
		v, err := auth.NewOIDCValidator(ctx, cfg.Auth.OIDC)
		if err != nil {
			fmt.Printf("OIDC provider unreachable at startup: %v. OIDC validation will fail until provider is reachable.\n", err)
		} else {
			oidcValidator = v
			fmt.Printf("OIDC provider: %s\n", cfg.Auth.OIDC.IssuerURL)
			fmt.Printf("OIDC scope claim: %s\n", cfg.Auth.OIDC.ScopeClaim)
			if len(cfg.Auth.OIDC.AllowedDomains) > 0 {
				fmt.Printf("OIDC allowed domains: %v\n", cfg.Auth.OIDC.AllowedDomains)
			}
		}
	} else {
		fmt.Println("OIDC: disabled")
	}

	rt := runtime.New(runtime.Options{
		Addr:            listenAddr,
		Version:         version,
		ShutdownTimeout: *grace,
		Provider:        provider,
		APIKeyStore:     keyStore,
		AuthConfig:      cfg.Auth,
		OIDCValidator:   oidcValidator,
	})
	inst, err := rt.Start()
	if err != nil {
		return err
	}

	fmt.Printf("Runtime listening on http://%s\n", inst.Addr())

	errCh := make(chan error, 1)
	go func() {
		errCh <- inst.Wait()
	}()

	select {
	case <-ctx.Done():
		shutdownErr := inst.Shutdown(context.Background())
		waitErr := <-errCh
		if shutdownErr != nil {
			return shutdownErr
		}
		if waitErr != nil {
			return waitErr
		}
		fmt.Println("Runtime stopped.")
		return nil
	case waitErr := <-errCh:
		if waitErr != nil {
			return waitErr
		}
		return nil
	}
}
