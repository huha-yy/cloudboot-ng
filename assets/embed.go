package assets

import (
	"embed"
	"io/fs"
)

// Embed web assets (templates and static files)
//
//go:embed all:web
var embeddedFiles embed.FS

// GetFS returns the embedded filesystem root
func GetFS() embed.FS {
	return embeddedFiles
}

// GetTemplatesFS returns only the templates subdirectory
func GetTemplatesFS() (fs.FS, error) {
	return fs.Sub(embeddedFiles, "web/templates")
}

// GetStaticFS returns only the static files subdirectory
func GetStaticFS() (fs.FS, error) {
	return fs.Sub(embeddedFiles, "web/static")
}
