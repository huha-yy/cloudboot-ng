package renderer

import (
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"path/filepath"

	"github.com/labstack/echo/v4"
)

// TemplateRenderer is a custom HTML template renderer for Echo
type TemplateRenderer struct {
	pages     map[string]*template.Template  // Independent template for each page
	funcMap   template.FuncMap
}

// NewTemplateRenderer creates a new template renderer from filesystem path
// This is used for development mode where templates are loaded from disk
func NewTemplateRenderer(templatesPath string) (*TemplateRenderer, error) {
	// Template functions
	funcMap := template.FuncMap{
		"sub": func(a, b int) int { return a - b },
		"add": func(a, b int) int { return a + b },
		"mul": func(a, b int) int { return a * b },
		"div": func(a, b int) int { if b == 0 { return 0 }; return a / b },
		"eq":  func(a, b interface{}) bool { return a == b },
		"len": func(v interface{}) int {
			switch val := v.(type) {
			case []interface{}:
				return len(val)
			case string:
				return len(val)
			default:
				return 0
			}
		},
	}

	renderer := &TemplateRenderer{
		pages:   make(map[string]*template.Template),
		funcMap: funcMap,
	}

	// Parse each page as an independent template set
	pagesPath := filepath.Join(templatesPath, "pages", "*.html")
	pageFiles, err := filepath.Glob(pagesPath)
	if err != nil {
		return nil, err
	}

	// Also load boot templates (Kickstart, iPXE, AutoYaST)
	bootPath := filepath.Join(templatesPath, "boot", "*.tmpl")
	bootFiles, err := filepath.Glob(bootPath)
	if err == nil {
		// Append boot templates to pages
		pageFiles = append(pageFiles, bootFiles...)
	}

	baseLayout := filepath.Join(templatesPath, "layouts", "base.html")
	componentsPattern := filepath.Join(templatesPath, "components", "*.html")

	for _, pageFile := range pageFiles {
		pageName := filepath.Base(pageFile)

		// Create a new template set for this page
		tmpl := template.New(pageName).Funcs(funcMap)

		// Boot templates (.tmpl) are standalone, don't need layout/components
		if filepath.Ext(pageFile) == ".tmpl" {
			tmpl, err = tmpl.ParseFiles(pageFile)
			if err != nil {
				return nil, fmt.Errorf("failed to parse boot template %s: %w", pageName, err)
			}
			renderer.pages[pageName] = tmpl
			continue
		}

		// Parse components (only for HTML pages)
		tmpl, err = tmpl.ParseGlob(componentsPattern)
		if err != nil {
			return nil, fmt.Errorf("failed to parse components for %s: %w", pageName, err)
		}

		// Parse base layout (only for HTML pages)
		tmpl, err = tmpl.ParseFiles(baseLayout)
		if err != nil {
			return nil, fmt.Errorf("failed to parse base layout for %s: %w", pageName, err)
		}

		// Parse the page file
		tmpl, err = tmpl.ParseFiles(pageFile)
		if err != nil {
			return nil, fmt.Errorf("failed to parse page %s: %w", pageName, err)
		}

		renderer.pages[pageName] = tmpl
	}

	return renderer, nil
}

// NewTemplateRendererFromFS creates a new template renderer from embed.FS
// This is used for production mode where templates are embedded in the binary
func NewTemplateRendererFromFS(templateFS fs.FS) (*TemplateRenderer, error) {
	// Template functions
	funcMap := template.FuncMap{
		"sub": func(a, b int) int { return a - b },
		"add": func(a, b int) int { return a + b },
		"mul": func(a, b int) int { return a * b },
		"div": func(a, b int) int { if b == 0 { return 0 }; return a / b },
		"eq":  func(a, b interface{}) bool { return a == b },
		"len": func(v interface{}) int {
			switch val := v.(type) {
			case []interface{}:
				return len(val)
			case string:
				return len(val)
			default:
				return 0
			}
		},
	}

	renderer := &TemplateRenderer{
		pages:   make(map[string]*template.Template),
		funcMap: funcMap,
	}

	// Get all page files
	pageFiles, err := fs.Glob(templateFS, "pages/*.html")
	if err != nil {
		return nil, err
	}

	// Get base layout
	baseContent, err := fs.ReadFile(templateFS, "layouts/base.html")
	if err != nil {
		return nil, fmt.Errorf("failed to read base layout: %w", err)
	}

	// Get all component files
	componentFiles, err := fs.Glob(templateFS, "components/*.html")
	if err != nil {
		return nil, err
	}

	// For each page, create an independent template set
	for _, pageFile := range pageFiles {
		pageName := filepath.Base(pageFile)

		// Create a new template set
		tmpl := template.New(pageName).Funcs(funcMap)

		// Parse components
		for _, compFile := range componentFiles {
			content, err := fs.ReadFile(templateFS, compFile)
			if err != nil {
				return nil, fmt.Errorf("failed to read component %s: %w", compFile, err)
			}
			_, err = tmpl.Parse(string(content))
			if err != nil {
				return nil, fmt.Errorf("failed to parse component %s: %w", compFile, err)
			}
		}

		// Parse base layout
		_, err = tmpl.Parse(string(baseContent))
		if err != nil {
			return nil, fmt.Errorf("failed to parse base layout for %s: %w", pageName, err)
		}

		// Parse the page file
		pageContent, err := fs.ReadFile(templateFS, pageFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read page %s: %w", pageFile, err)
		}
		_, err = tmpl.Parse(string(pageContent))
		if err != nil {
			return nil, fmt.Errorf("failed to parse page %s: %w", pageName, err)
		}

		renderer.pages[pageName] = tmpl
	}

	return renderer, nil
}

// Render renders a template with the given data
func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	// Look up the page-specific template set
	tmpl, ok := t.pages[name]
	if !ok {
		return fmt.Errorf("template %s not found", name)
	}

	// Boot templates (.tmpl) are standalone - execute directly
	if filepath.Ext(name) == ".tmpl" {
		return tmpl.Execute(w, data)
	}

	// HTML pages execute base.html which will call the content block
	// The content block is defined in each page template as {{define "content"}}
	return tmpl.ExecuteTemplate(w, "base.html", data)
}
