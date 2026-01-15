package web

import (
	"embed"
	"io/fs"
)

// StaticFiles embeds all static resources (CSS, JS, images, etc.)
//
//go:embed static
var StaticFiles embed.FS

// TemplateFiles embeds all HTML templates
//
//go:embed templates
var TemplateFiles embed.FS

// GetStaticAssets returns the static file system without the "static/" prefix
// This allows serving files directly without the path prefix
// Example: /static/css/output.css -> /css/output.css
func GetStaticAssets() (fs.FS, error) {
	return fs.Sub(StaticFiles, "static")
}

// GetTemplateAssets returns the template file system without the "templates/" prefix
// This is useful for template engines that expect templates at the root
func GetTemplateAssets() (fs.FS, error) {
	return fs.Sub(TemplateFiles, "templates")
}

// GetRawStaticFS returns the raw embedded static file system (with "static/" prefix)
// Use this if you need access to the full path structure
func GetRawStaticFS() fs.FS {
	return StaticFiles
}

// GetRawTemplateFS returns the raw embedded template file system (with "templates/" prefix)
// Use this if you need access to the full path structure
func GetRawTemplateFS() fs.FS {
	return TemplateFiles
}
