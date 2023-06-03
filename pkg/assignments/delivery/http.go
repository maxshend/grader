package delivery

import (
	"fmt"
	"io/fs"
	"net/http"
	"strconv"
	"strings"

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

type newAssignmentnData struct {
	Assignment *assignments.Assignment
	Files      string
	Errors     []string
	Action     string
}

func NewAssignmentsHttpHandler(
	service services.AssignmentsServiceInterface,
	sessionManager sessions.HttpSessionManager,
	ubmissionsService submissionsServices.SubmissionsServiceInterface,
	templatesFS fs.FS,
) (*AssignmentsHttpHandler, error) {
	views := make(map[string]*utils.View)
	var err error

	views["GetAll"], err = utils.NewView(templatesFS, "templates/assignments/admin/list.gohtml")
	if err != nil {
		return nil, err
	}
	views["AssignmentForm"], err = utils.NewView(templatesFS, "templates/assignments/admin/assignment_form.gohtml")
	if err != nil {
		return nil, err
	}
	views["Show"], err = utils.NewView(templatesFS, "templates/assignments/admin/assignment.gohtml")
	if err != nil {
		return nil, err
	}

	views["GetAllPersonal"], err = utils.NewView(templatesFS, "templates/assignments/list.gohtml")
	if err != nil {
		return nil, err
	}
	views["NewSubmission"], err = utils.NewView(templatesFS, "templates/assignments/new_submission.gohtml")
	if err != nil {
		return nil, err
	}
	views["ShowPersonal"], err = utils.NewView(templatesFS, "templates/assignments/assignment.gohtml")
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
	result, err := h.Service.GetAll(currentUser)
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

	assignment, err := h.Service.GetByID(assignmentID(params["id"]))
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

	assignment, err := h.Service.GetByID(assignmentID(params["id"]))
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
	assignment, err := h.Service.GetByID(assignmentID(params["id"]))
	if err != nil {
		utils.RenderInternalError(w, r, err)
		return
	}
	if assignment == nil {
		http.NotFound(w, r)
		return
	}

	page := utils.GetPageNumber(r)
	submissionsList, paginationData, err := h.SubmissionsService.GetByUserAssignment(
		assignment.ID,
		currentUser.ID,
		page,
	)
	if err != nil {
		utils.RenderInternalError(w, r, err)
		return
	}

	err = h.Views["ShowPersonal"].RenderView(
		w,
		&struct {
			Assignment     *assignments.Assignment
			Submissions    []*submissions.Submission
			PaginationData *utils.PaginationData
		}{assignment, submissionsList, paginationData},
		currentUser,
	)
	if err != nil {
		utils.RenderInternalError(w, r, err)
	}
}

func (h AssignmentsHttpHandler) New(w http.ResponseWriter, r *http.Request) {
	currentUser, err := h.SessionManager.CurrentUser(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	err = h.Views["AssignmentForm"].RenderView(
		w,
		newAssignmentnData{
			Assignment: &assignments.Assignment{},
			Action:     "create",
		},
		currentUser,
	)
	if err != nil {
		utils.RenderInternalError(w, r, err)
	}
}

func (h AssignmentsHttpHandler) Create(w http.ResponseWriter, r *http.Request) {
	currentUser, err := h.SessionManager.CurrentUser(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	assignment := &assignments.Assignment{
		CreatorID:   currentUser.ID,
		Title:       r.FormValue("title"),
		Description: r.FormValue("description"),
		GraderURL:   r.FormValue("grader_url"),
		Container:   r.FormValue("container"),
		PartID:      r.FormValue("part_id"),
		Files:       formatAssignmentFiles(r.FormValue("files")),
	}
	_, err = h.Service.Create(assignment)
	if err != nil {
		if _, ok := err.(*services.AssignmentValidationError); ok {
			err = h.Views["AssignmentForm"].RenderView(
				w,
				newAssignmentnData{
					Assignment: assignment,
					Files:      r.FormValue("files"),
					Errors:     []string{err.Error()},
					Action:     "create",
				},
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

	http.Redirect(w, r, "/admin/assignments", http.StatusSeeOther)
}

func (h AssignmentsHttpHandler) Edit(w http.ResponseWriter, r *http.Request) {
	currentUser, err := h.SessionManager.CurrentUser(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	params := mux.Vars(r)
	assignment, err := h.Service.GetByIDByCreator(assignmentID(params["id"]), currentUser)
	if err != nil {
		utils.RenderInternalError(w, r, err)
		return
	}
	if assignment == nil {
		http.NotFound(w, r)
		return
	}

	err = h.Views["AssignmentForm"].RenderView(
		w,
		newAssignmentnData{
			Assignment: assignment,
			Files:      strings.Join(assignment.Files, ","),
			Action:     "update",
		},
		currentUser,
	)
	if err != nil {
		utils.RenderInternalError(w, r, err)
	}
}

func (h AssignmentsHttpHandler) Update(w http.ResponseWriter, r *http.Request) {
	currentUser, err := h.SessionManager.CurrentUser(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	params := mux.Vars(r)
	assignment, err := h.Service.GetByIDByCreator(assignmentID(params["id"]), currentUser)
	if err != nil {
		utils.RenderInternalError(w, r, err)
		return
	}
	if assignment == nil {
		http.NotFound(w, r)
		return
	}

	assignment.Title = r.FormValue("title")
	assignment.Description = r.FormValue("description")
	assignment.GraderURL = r.FormValue("grader_url")
	assignment.Container = r.FormValue("container")
	assignment.PartID = r.FormValue("part_id")
	assignment.Files = formatAssignmentFiles(r.FormValue("files"))

	_, err = h.Service.Update(assignment)
	if err != nil {
		if _, ok := err.(*services.AssignmentValidationError); ok {
			err = h.Views["AssignmentForm"].RenderView(
				w,
				newAssignmentnData{
					Assignment: assignment,
					Files:      r.FormValue("files"),
					Errors:     []string{err.Error()},
					Action:     "update",
				},
				currentUser,
			)
			if err != nil {
				utils.RenderInternalError(w, r, err)
			}
		} else {
			utils.RenderInternalError(w, r, err)
		}
	}

	http.Redirect(w, r, "/admin/assignments", http.StatusSeeOther)
}

func (h AssignmentsHttpHandler) Show(w http.ResponseWriter, r *http.Request) {
	currentUser, err := h.SessionManager.CurrentUser(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	params := mux.Vars(r)
	assignment, err := h.Service.GetByIDByCreator(assignmentID(params["id"]), currentUser)
	if err != nil {
		utils.RenderInternalError(w, r, err)
		return
	}
	if assignment == nil {
		http.NotFound(w, r)
		return
	}
	submissionsList, err := h.SubmissionsService.GetByAssignment(assignment.ID)
	if err != nil {
		utils.RenderInternalError(w, r, err)
		return
	}

	err = h.Views["Show"].RenderView(
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

func formatAssignmentFiles(files string) []string {
	return strings.Split(files, ",")
}

func assignmentID(param string) int64 {
	id, _ := strconv.ParseInt(param, 10, 64)

	return id
}
