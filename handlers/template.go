package handlers

import (
	"path/filepath"
	"net/http"
	"github.com/stretchr/objx"
	"html/template"
)

type templateHandler struct {
	filename string
	templ *template.Template
}

func NewTemplateHandler(filename string) *templateHandler {
	return &templateHandler{filename, template.Must(template.ParseFiles(filepath.Join("html", filename)))}
}

func (t *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request){
	data := map[string]interface{}{
		"Host": r.Host,
	}
	if authCookie, err := r.Cookie("auth"); err == nil {
		data["UserData"] = objx.MustFromBase64(authCookie.Value)
	}
	t.templ.Execute(w, data)
}
