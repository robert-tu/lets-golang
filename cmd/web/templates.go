package main

import (
	"html/template"
	"path/filepath"
	"time"

	"robert-tu.net/snippetbox/pkg/forms"
	"robert-tu.net/snippetbox/pkg/models"
)

// define templateData type struct
type templateData struct {
	Snippet         *models.Snippet
	Snippets        []*models.Snippet
	CurrentYear     int
	Form            *forms.Form
	Flash           string
	IsAuthenticated bool
	CSRFToken       string
}

// humanDate function returning formatted date
func humanDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format("Jan 02 2006 at 15:04")
}

// initialize template.FuncMap as global variable
var functions = template.FuncMap{
	"humanDate": humanDate,
}

// define newTemplateCache function
func newTemplateCache(dir string) (map[string]*template.Template, error) {
	// initialize map
	cache := map[string]*template.Template{}

	// get slice of all filepaths
	pages, err := filepath.Glob(filepath.Join(dir, "*.page.tmpl"))
	if err != nil {
		return nil, err
	}

	// loop pages
	for _, page := range pages {
		// extract file and assign
		name := filepath.Base(page)
		// parse into template set
		ts, err := template.New(name).Funcs(functions).ParseFiles(page)
		if err != nil {
			return nil, err
		}
		// add layout templates to template set
		ts, err = ts.ParseGlob(filepath.Join(dir, "*.layout.tmpl"))
		if err != nil {
			return nil, err
		}
		// add partial templates to template set
		ts, err = ts.ParseGlob(filepath.Join(dir, "*.partial.tmpl"))
		if err != nil {
			return nil, err
		}
		// add template set to cache
		cache[name] = ts
	}
	return cache, nil
}
