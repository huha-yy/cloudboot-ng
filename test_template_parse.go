package main

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
)

//go:embed web/templates
var templateFS embed.FS

func main() {
	pageFiles, _ := fs.Glob(templateFS, "pages/*.html")
	baseContent, _ := fs.ReadFile(templateFS, "layouts/base.html")
	componentFiles, _ := fs.Glob(templateFS, "components/*.html")

	for _, pageFile := range pageFiles {
		pageName := pageFile
		tmpl := template.New(pageName)

		// Parse components
		for _, compFile := range componentFiles {
			content, _ := fs.ReadFile(templateFS, compFile)
			_, err := tmpl.Parse(string(content))
			if err != nil {
				fmt.Printf("Component %s error: %v\n", compFile, err)
			}
		}

		// Parse base layout
		_, err := tmpl.Parse(string(baseContent))
		if err != nil {
			fmt.Printf("Base layout error for %s: %v\n", pageName, err)
			continue
		}

		// Parse page file
		pageContent, _ := fs.ReadFile(templateFS, pageFile)
		_, err = tmpl.Parse(string(pageContent))
		if err != nil {
			fmt.Printf("Page %s error: %v\n", pageName, err)
		} else {
			fmt.Printf("Page %s: OK\n", pageName)
		}
	}
}
