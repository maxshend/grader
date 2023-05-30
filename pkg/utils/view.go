package utils

import (
	"html/template"
	"io/fs"
	"net/http"

	"github.com/maxshend/grader/pkg/submissions"
	"github.com/maxshend/grader/pkg/users"
)

type View struct {
	Template *template.Template
	Layout   string
}

func NewView(templatesFS fs.FS, files ...string) (*View, error) {
	layouts, err := layoutFiles(templatesFS)
	if err != nil {
		return nil, err
	}

	name := files[0]
	files = append(layouts, files...)

	t, err := template.New(name).Funcs(
		template.FuncMap{
			"currentUser":     func() *users.User { return nil },
			"isAuthenticated": func() bool { return false },
			"submissionStatus": func(status int) string {
				switch status {
				case submissions.InProgress:
					return "Waiting"
				case submissions.Success:
					return "Success"
				case submissions.Fail:
					return "Fail"
				}

				return "Unknown"
			},
			"userProvider": func(provider int) string {
				switch provider {
				case users.DefaultProvider:
					return "Username"
				case users.VkProvider:
					return "VK"
				}

				return "Unknown"
			},
		},
	).ParseFS(templatesFS, files...)
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

func layoutFiles(templatesFS fs.FS) ([]string, error) {
	files, err := fs.Glob(templatesFS, "templates/layouts/*.gohtml")
	if err != nil {
		return nil, err
	}

	return files, nil
}
