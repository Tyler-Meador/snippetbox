package main

import (
	"html/template"
	"io/fs"
	"path/filepath"
	"time"

	"github.com/Tyler-Meador/snippetbox/internal/models"
	"github.com/Tyler-Meador/snippetbox/ui"
)

type templateData struct {
	Snippet         models.Snippet
	Snippets        []models.Snippet
	CurrentYear     int
	Form            any
	Flash           string
	IsAuthenticated bool
	CSRFToken       string
}

var functions = template.FuncMap{
	"humanDate": humanDate,
}

func newTemplateCache() (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}

	pages, err := fs.Glob(ui.Files, "html/pages/*.html")
	if err != nil {
		return nil, err
	}

	for _, page := range pages {
		name := filepath.Base(page)

		patterns := []string{
			"html/base.html",
			"html/partials/*.html",
			page,
		}

		tmps, err := template.New(name).Funcs(functions).ParseFS(ui.Files, patterns...)
		if err != nil {
			return nil, err
		}

		cache[name] = tmps
	}

	return cache, nil
}

func humanDate(humanTime time.Time) string {
	return humanTime.Format("02 Jan 2006 at 15:04")
}
