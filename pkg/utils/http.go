package utils

import (
	"math"
	"net/http"
	"strconv"
)

type PaginationData struct {
	CurrentPage int
	MaxPage     int
	PrevPage    int
	NextPage    int
	LastPage    bool
	FirstPage   bool
}

func RedirectUnauthenticated(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/signin", http.StatusSeeOther)
}

func RenderInternalError(w http.ResponseWriter, r *http.Request, err error) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

func BoolFromParam(value string) bool {
	return len(value) != 0
}

func GetPageNumber(r *http.Request) int {
	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil {
		return 1
	}

	return page
}

func GetMaxPage(pageSize, totalCount int) int {
	return int(math.Ceil(float64(totalCount) / float64(pageSize)))
}

func GetPageOffset(page, pageSize int) int {
	return (page - 1) * pageSize
}
