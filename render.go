package main

import (
	"bytes"
	"html/template"
	"net/http"

	"github.com/gorilla/csrf"
)

// render executes a template registered under name as the inner content of
// templates/layout.html. It clones the parsed template per request so it can
// bind {{ csrfField }} and {{ csrfToken }} to the current request's CSRF token.
//
// Output is buffered so that template execution errors do not produce a partial
// response (which would also have already sent a 200 status).
func (a *App) render(w http.ResponseWriter, r *http.Request, name string, data PageData) {
	tmpl, ok := a.Templates[name]
	if !ok || tmpl == nil {
		http.Error(w, "template not found: "+name, http.StatusInternalServerError)
		return
	}
	clone, err := tmpl.Clone()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	clone = clone.Funcs(template.FuncMap{
		"csrfField": func() template.HTML { return csrf.TemplateField(r) },
		"csrfToken": func() string { return csrf.Token(r) },
	})

	var buf bytes.Buffer
	if err := clone.ExecuteTemplate(&buf, "layout.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, _ = w.Write(buf.Bytes())
}
