// Package main is the entry point for the server application.
package main

import (
	"github.com/masmuss/gokit-starter/internal/app"
	"go.uber.org/fx"
)

func main() {
	fx.New(app.Module).Run()
}
