// Package main is the entry point for the server application.
package main

import (
	"fmt"
	"os"

	"github.com/masmuss/gokit-starter/internal/app"
	"go.uber.org/fx"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "viz" {
		// Newer Fx uses fx.DotGraph to get the graph as a DOT string
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
