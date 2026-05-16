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
	// Add fields here
}

// Repository defines the interface for {{.Name}} data persistence.
type Repository interface {
	// Add repository methods here
}
`

const appTemplate = `package app

import (
	"github.com/masmuss/gokit-starter/internal/modules/{{.Name}}/domain"
)

// Service implements the use cases for {{.Name}}.
type Service struct {
	repo domain.Repository
}

// NewService creates a new {{.Name}} service.
func NewService(repo domain.Repository) *Service {
	return &Service{
		repo: repo,
	}
}
`

const infraTemplate = `package infra

import (
	"github.com/masmuss/gokit-starter/internal/database/ent"
	"github.com/masmuss/gokit-starter/internal/modules/{{.Name}}/domain"
)

// Repository implements domain.Repository using Ent.
type Repository struct {
	client *ent.Client
}

// NewRepository creates a new {{.Name}} repository.
func NewRepository(client *ent.Client) *Repository {
	return &Repository{
		client: client,
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
		filepath.Join(basePath, "infra", "repository.go"): infraTemplate,
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
	fmt.Printf("1. Define your schema in internal/database/schema/%s.go\n", moduleName)
	fmt.Println("2. Run 'task generate'")
	fmt.Printf("3. Register the new module in internal/app/fx.go\n")
}
