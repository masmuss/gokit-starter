// Package main provides a module generator for scaffolding new business modules.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"unicode"
)

type ModuleData struct {
	Name      string
	TitleName string
}

const domainTemplate = `package domain

// {{.TitleName}} represents the domain entity.
type {{.TitleName}} struct {
	ID string ` + "`" + `json:"id"` + "`" + `
}
`

const appTemplate = `package app

import (
	"context"

	"github.com/masmuss/gokit-starter/internal/modules/{{.Name}}/domain"
)

// Repository defines the interface for {{.Name}} data persistence.
type Repository interface {
	FindByID(ctx context.Context, id any) (domain.{{.TitleName}}, error)
}

// Service implements the use cases for {{.Name}}.
type Service struct {
	repo Repository
}

// NewService creates a new {{.Name}} service.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}
`

const handlerTemplate = `package handler

import (
	"log/slog"
	"net/http"

	chi "github.com/go-chi/chi/v5"

	"github.com/masmuss/gokit-starter/internal/delivery/response"
)

// Service defines the interface for {{.Name}} operations.
type Service interface {
}

// Handler handles {{.Name}} requests.
type Handler struct {
	service Service
	log     *slog.Logger
}

// NewHandler creates a new {{.Name}} handler.
func NewHandler(service Service, log *slog.Logger) *Handler {
	return &Handler{service: service, log: log}
}

// RegisterRoutes registers {{.Name}} routes.
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/{{.Name}}", func(r chi.Router) {
	})
}
`

const infraTemplate = `package infra

import (
	"gorm.io/gorm"

	"github.com/masmuss/gokit-starter/internal/infra/database"
	"github.com/masmuss/gokit-starter/internal/modules/{{.Name}}/app"
)

// Repository implements app.Repository using GORM.
type Repository struct {
	db *gorm.DB
}

// NewRepository creates a new {{.Name}} repository.
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// NewRepositoryFromDB creates a Repository from database.DB.
func NewRepositoryFromDB(database *database.DB) *Repository {
	return NewRepository(database.DB)
}
`

const wireTemplate = `package {{.Name}}

import (
	chi "github.com/go-chi/chi/v5"

	"github.com/masmuss/gokit-starter/internal/delivery"
	"github.com/masmuss/gokit-starter/internal/modules/{{.Name}}/handler"
)

// Module exposes the {{.Name}} module's public outputs.
type Module struct {
	Handler   *handler.Handler
	Registrar delivery.RouteRegistrar
}

// Wire builds the {{.Name}} module from its dependencies.
func Wire() Module {
	h := handler.NewHandler(nil, nil) // TODO: inject real dependencies
	return Module{
		Handler: h,
		Registrar: delivery.RouteRegistrarFunc(func(r chi.Router) {
			h.RegisterRoutes(r)
		}),
	}
}
`

func toTitle(s string) string {
	if s == "" {
		return ""
	}
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run scripts/generator/main.go <module_name>")
		os.Exit(1)
	}

	moduleName := strings.ToLower(os.Args[1])
	data := ModuleData{
		Name:      moduleName,
		TitleName: toTitle(moduleName),
	}

	basePath := filepath.Join("internal", "modules", moduleName)

	folders := []string{
		filepath.Join(basePath, "domain"),
		filepath.Join(basePath, "app"),
		filepath.Join(basePath, "handler"),
		filepath.Join(basePath, "infra"),
	}

	for _, folder := range folders {
		if err := os.MkdirAll(folder, 0755); err != nil {
			fmt.Printf("Error creating folder %s: %v\n", folder, err)
			os.Exit(1)
		}
	}

	files := map[string]string{
		filepath.Join(basePath, "domain", "model.go"):     domainTemplate,
		filepath.Join(basePath, "app", "service.go"):      appTemplate,
		filepath.Join(basePath, "handler", "handler.go"):  handlerTemplate,
		filepath.Join(basePath, "infra", "repository.go"): infraTemplate,
		filepath.Join(basePath, "wire.go"):                wireTemplate,
	}

	for path, content := range files {
		tmpl, err := template.New("tmpl").Parse(content)
		if err != nil {
			fmt.Printf("Error parsing template for %s: %v\n", path, err)
			os.Exit(1)
		}

		func() {
			f, createErr := os.Create(path)
			if createErr != nil {
				fmt.Printf("Error creating file %s: %v\n", path, createErr)
				os.Exit(1)
			}
			defer f.Close()

			if execErr := tmpl.Execute(f, data); execErr != nil {
				fmt.Printf("Error executing template for %s: %v\n", path, execErr)
				os.Exit(1)
			}
		}()
		fmt.Printf("Created: %s\n", path)
	}

	fmt.Printf("\nModule '%s' scaffolded successfully!\n", moduleName)
	fmt.Println("Next steps:")
	fmt.Printf("1. Define your model in internal/database/model/%s.go\n", moduleName)
	fmt.Println("2. Add model to AutoMigrate in internal/infra/database/database.go")
	fmt.Printf("3. Call {{.Name}}.Wire() in cmd/server/main.go\n")
}
