package delivery

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/maxshend/grader/pkg/assignments"
	"github.com/maxshend/grader/pkg/assignments/services"
	"github.com/maxshend/grader/pkg/sessions"
	"github.com/maxshend/grader/pkg/submissions"
	submissionsServices "github.com/maxshend/grader/pkg/submissions/services"
	"github.com/maxshend/grader/pkg/utils"
)

type AssignmentsHttpHandler struct {
	Service            services.AssignmentsServiceInterface
	SubmissionsService submissionsServices.SubmissionsServiceInterface
	SessionManager     sessions.HttpSessionManager
	Views              map[string]*utils.View
}

type newSubmissionData struct {
	Assignment *assignments.Assignment
	Errors     []string
}

func NewAssignmentsHttpHandler(
	service services.AssignmentsServiceInterface,
	sessionManager sessions.HttpSessionManager,
	ubmissionsService submissionsServices.SubmissionsServiceInterface,
) (*AssignmentsHttpHandler, error) {
	views := make(map[string]*utils.View)
	var err error

	views["GetAll"], err = utils.NewView("./web/templates/assignments/admin/list.gohtml")
	if err != nil {
		return nil, err
	}
	views["GetAllPersonal"], err = utils.NewView("./web/templates/assignments/list.gohtml")
	if err != nil {
		return nil, err
	}
	views["NewSubmission"], err = utils.NewView("./web/templates/assignments/new_submission.gohtml")
	if err != nil {
		return nil, err
	}
	views["ShowPersonal"], err = utils.NewView("./web/templates/assignments/assignment.gohtml")
	if err != nil {
		return nil, err
	}

	return &AssignmentsHttpHandler{
		Service:            service,
		Views:              views,
		SessionManager:     sessionManager,
		SubmissionsService: ubmissionsService,
	}, nil
}

func (h AssignmentsHttpHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	currentUser, err := h.SessionManager.CurrentUser(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	result, err := h.Service.GetAll()
	if err != nil {
		utils.RenderInternalError(w, r, err)
		return
	}

	err = h.Views["GetAll"].RenderView(
		w,
		&struct{ Assignments []*assignments.Assignment }{result},
		currentUser,
	)
	if err != nil {
		utils.RenderInternalError(w, r, err)
	}
}

func (h AssignmentsHttpHandler) PersonalAssignments(w http.ResponseWriter, r *http.Request) {
	currentUser, err := h.SessionManager.CurrentUser(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	result, err := h.Service.GetByUserID(currentUser.ID)
	if err != nil {
		utils.RenderInternalError(w, r, err)
		return
	}

	err = h.Views["GetAllPersonal"].RenderView(
		w,
		&struct{ Assignments []*assignments.Assignment }{result},
		currentUser,
	)
	if err != nil {
		utils.RenderInternalError(w, r, err)
	}
}

func (h AssignmentsHttpHandler) NewSubmission(w http.ResponseWriter, r *http.Request) {
	currentUser, err := h.SessionManager.CurrentUser(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	params := mux.Vars(r)

	assignment, err := h.Service.GetByID(params["id"])
	if err != nil {
		utils.RenderInternalError(w, r, err)
		return
	}
	if assignment == nil {
		http.NotFound(w, r)
		return
	}

	err = h.Views["NewSubmission"].RenderView(
		w,
		&newSubmissionData{Assignment: assignment},
		currentUser,
	)
	if err != nil {
		utils.RenderInternalError(w, r, err)
	}
}

func (h AssignmentsHttpHandler) CreateSubmission(w http.ResponseWriter, r *http.Request) {
	currentUser, err := h.SessionManager.CurrentUser(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 5*1024*1024)
	params := mux.Vars(r)

	assignment, err := h.Service.GetByID(params["id"])
	if err != nil {
		utils.RenderInternalError(w, r, err)
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
			utils.RenderInternalError(w, r, err)
			return
		}
		defer uploadData.Close()

		files = append(files, &services.SubmissionFile{Content: uploadData, Name: header.Filename})
	}

	_, err = h.Service.Submit(currentUser, assignment, files)
	if err != nil {
		if _, ok := err.(*services.AssignmentValidationError); ok {
			err = h.Views["NewSubmission"].RenderView(
				w,
				newSubmissionData{assignment, []string{err.Error()}},
				currentUser,
			)
			if err != nil {
				utils.RenderInternalError(w, r, err)
			}
		} else {
			utils.RenderInternalError(w, r, err)
		}

		return
	}

	http.Redirect(w, r, fmt.Sprintf("/assignments/%d", assignment.ID), http.StatusSeeOther)
}

func (h AssignmentsHttpHandler) ShowPersonal(w http.ResponseWriter, r *http.Request) {
	currentUser, err := h.SessionManager.CurrentUser(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	params := mux.Vars(r)
	assignment, err := h.Service.GetByID(params["id"])
	if err != nil {
		utils.RenderInternalError(w, r, err)
		return
	}
	if assignment == nil {
		http.NotFound(w, r)
		return
	}
	submissionsList, err := h.SubmissionsService.GetByUserAssignment(assignment.ID, currentUser.ID)
	if err != nil {
		utils.RenderInternalError(w, r, err)
		return
	}

	err = h.Views["ShowPersonal"].RenderView(
		w,
		&struct {
			Assignment  *assignments.Assignment
			Submissions []*submissions.Submission
		}{assignment, submissionsList},
		currentUser,
	)
	if err != nil {
		utils.RenderInternalError(w, r, err)
	}
}
