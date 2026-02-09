package backend

import "net/http"

type PageState int

const (
	ServerManagement PageState = iota
)

var CurrentPage PageState = ServerManagement

func CurrentPageHandler(w http.ResponseWriter, r *http.Request) {
	address := []string{"/templates/server_managing.html"}[CurrentPage]
	http.Redirect(w, r, address, http.StatusSeeOther)
}
