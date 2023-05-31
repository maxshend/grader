package utils

import "net/http"

func RedirectUnauthenticated(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/signin", http.StatusSeeOther)
}

func RenderInternalError(w http.ResponseWriter, r *http.Request, err error) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

func BoolFromParam(value string) bool {
	return len(value) != 0
}
