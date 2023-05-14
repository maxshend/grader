package utils

import (
	"html/template"
	"net/http"
	"path/filepath"
)

type View struct {
	Template *template.Template
	Layout   string
}

func NewView(files ...string) (*View, error) {
	layouts, err := layoutFiles()
	if err != nil {
		return nil, err
	}
	files = append(layouts, files...)

	t, err := template.ParseFiles(files...)
	if err != nil {
		panic(err)
	}

	return &View{
		Template: t,
	}, nil
}

func (v *View) RenderView(w http.ResponseWriter, data interface{}) error {
	return v.Template.ExecuteTemplate(w, "main", data)
}

func layoutFiles() ([]string, error) {
	files, err := filepath.Glob("./web/templates/layouts/*.gohtml")
	if err != nil {
		return nil, err
	}

	return files, nil
}
