package main

import (
	"html/template"
	"path/filepath"
	"time"

	"github.com/Tyler-Meador/snippetbox/internal/models"
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

	pages, err := filepath.Glob("./ui/html/pages/*.html")
	if err != nil {
		return nil, err
	}

	for _, page := range pages {
		name := filepath.Base(page)

		tmps, err := template.New(name).Funcs(functions).ParseFiles("./ui/html/base.html")
		if err != nil {
			return nil, err
		}

		tmps, err = tmps.ParseGlob("./ui/html/partials/*.html")
		if err != nil {
			return nil, err
		}

		tmps, err = tmps.ParseFiles(page)
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
