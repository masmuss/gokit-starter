// Package main is the entry point for the server application.
package main

import (
	"fmt"
	"os"

	"go.uber.org/fx"

	"github.com/masmuss/gokit-starter/internal/app"
)

// version is injected at build time via -ldflags.
// Example: go build -ldflags="-X main.version=v1.2.3" -o bin/app ./cmd/server
var version = "dev"

func main() {
	os.Setenv("APP_VERSION", version)

	if len(os.Args) > 1 && os.Args[1] == "viz" {
		var dot fx.DotGraph
		app := fx.New(
			app.Module,
			fx.Populate(&dot),
		)
		if err := app.Err(); err != nil {
			fmt.Fprintf(os.Stderr, "failed to build app graph: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(dot))
		return
	}

	fx.New(app.Module).Run()
}
