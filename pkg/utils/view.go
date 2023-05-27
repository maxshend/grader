package utils

import (
	"html/template"
	"net/http"
	"path/filepath"

	"github.com/maxshend/grader/pkg/users"
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

	name := files[0]
	files = append(layouts, files...)

	t, err := template.New(name).Funcs(
		template.FuncMap{
			"currentUser":     func() *users.User { return nil },
			"isAuthenticated": func() bool { return false },
		},
	).ParseFiles(files...)
	if err != nil {
		panic(err)
	}

	return &View{
		Template: t,
	}, nil
}

func (v *View) RenderView(w http.ResponseWriter, data interface{}, currentUser *users.User) error {
	return template.Must(v.Template.Clone()).Funcs(
		template.FuncMap{
			"currentUser":     func() *users.User { return currentUser },
			"isAuthenticated": func() bool { return currentUser != nil },
		},
	).ExecuteTemplate(w, "main", data)
}

func layoutFiles() ([]string, error) {
	files, err := filepath.Glob("./web/templates/layouts/*.gohtml")
	if err != nil {
		return nil, err
	}

	return files, nil
}
