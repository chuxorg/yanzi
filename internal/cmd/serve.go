package cmd

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/chuxorg/yanzi/internal/runtime"
)

// RunServe starts the shared runtime server in the foreground.
func RunServe(args []string, version string) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	return runServe(ctx, args, version)
}

func runServe(ctx context.Context, args []string, version string) error {
	fs := flag.NewFlagSet("serve", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	addr := fs.String("addr", "127.0.0.1:8080", "listen address")
	grace := fs.Duration("shutdown-timeout", 5*time.Second, "graceful shutdown timeout")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return errors.New("usage: yanzi serve [--addr host:port] [--shutdown-timeout duration]")
	}

	rt := runtime.New(runtime.Options{
		Addr:            strings.TrimSpace(*addr),
		Version:         version,
		ShutdownTimeout: *grace,
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
