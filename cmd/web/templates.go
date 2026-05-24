package main

import (
	"html/template"
	"path/filepath"

	"github.com/mdombrov-33/snippetbox/internal/models"
)

// Define a templateData type to act as the holding structure for any dynamic data that we want to pass to our HTML templates.
// We're doing this because html/template package allows to pass only one item of dynamic data when rendering a template. By creating custom struct, we create a holding structure for our data.
type templateData struct {
	Snippet  *models.Snippet
	Snippets []*models.Snippet
}

func newTemplateCache() (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}

	pages, err := filepath.Glob("./ui/html/pages/*.html")
	if err != nil {
		return nil, err
	}

	for _, page := range pages {
		name := filepath.Base(page)
		ts, err := template.ParseFiles("./ui/html/base.html")

		if err != nil {
			return nil, err
		}

		ts, err = ts.ParseGlob("./ui/html/partials/*.html")
		if err != nil {
			return nil, err
		}

		ts, err = ts.ParseFiles(page)
		if err != nil {
			return nil, err
		}

		cache[name] = ts
	}

	return cache, nil
}
