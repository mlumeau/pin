package admin

import (
	"net/http"
)

// Root handles the HTTP request.
func (h Handler) Root(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/settings/profile", http.StatusFound)
}
