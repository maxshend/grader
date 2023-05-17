package delivery

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/maxshend/grader/pkg/assignments"
	"github.com/maxshend/grader/pkg/assignments/services"
	"github.com/maxshend/grader/pkg/utils"
)

type AssignmentsHttpHandler struct {
	Service services.AssignmentsServiceInterface
	Views   map[string]*utils.View
}

type newSubmissionData struct {
	Assignment *assignments.Assignment
	Errors     []string
}

func NewAssignmentsHttpHandler(
	service services.AssignmentsServiceInterface,
) (*AssignmentsHttpHandler, error) {
	views := make(map[string]*utils.View)
	var err error

	views["GetAll"], err = utils.NewView("./web/templates/assignments/list.gohtml")
	if err != nil {
		return nil, err
	}
	views["NewSubmission"], err = utils.NewView("./web/templates/assignments/new_submission.gohtml")
	if err != nil {
		return nil, err
	}

	return &AssignmentsHttpHandler{
		Service: service,
		Views:   views,
	}, nil
}

func (h AssignmentsHttpHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	result, err := h.Service.GetAll()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = h.Views["GetAll"].RenderView(w, &struct{ Assignments []*assignments.Assignment }{result})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h AssignmentsHttpHandler) NewSubmission(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	assignment, err := h.Service.GetByID(params["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if assignment == nil {
		http.NotFound(w, r)
		return
	}

	err = h.Views["NewSubmission"].RenderView(w, &newSubmissionData{Assignment: assignment})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h AssignmentsHttpHandler) Submit(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 5*1024*1024)
	params := mux.Vars(r)

	assignment, err := h.Service.GetByID(params["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if assignment == nil {
		http.NotFound(w, r)
		return
	}

	files := []*services.SubmissionFile{}
	for _, filename := range assignment.Files {
		uploadData, header, err := r.FormFile(filename)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer uploadData.Close()

		files = append(files, &services.SubmissionFile{Content: uploadData, Name: header.Filename})
	}

	_, err = h.Service.Submit(assignment, files)
	if err != nil {
		if _, ok := err.(*services.AssignmentValidationError); ok {
			err = h.Views["NewSubmission"].RenderView(
				w,
				newSubmissionData{assignment, []string{err.Error()}},
			)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		return
	}

	http.Redirect(w, r, fmt.Sprintf("/assignments/%d", assignment.ID), http.StatusSeeOther)
}
