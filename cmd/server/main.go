// Package main starts the HTTP server.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/masmuss/gokit-starter/internal/app"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, log, db, err := app.Bootstrap(ctx, os.Stdout)
	if err != nil {
		return err
	}

	application := app.New(cfg, log, db)

	if runErr := application.Start(ctx); runErr != nil {
		return runErr
	}

	return nil
}
