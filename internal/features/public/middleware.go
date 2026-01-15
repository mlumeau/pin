package public

import (
	"context"
	"net/http"
	"strings"
)

type SetupRedirectDependencies interface {
	HasUser(ctx context.Context) (bool, error)
}

// WithSetupRedirect sends first-time visitors to the setup page.
func WithSetupRedirect(deps SetupRedirectDependencies, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/setup" || strings.HasPrefix(r.URL.Path, "/invite/") || strings.HasPrefix(r.URL.Path, "/static/") {
			next.ServeHTTP(w, r)
			return
		}
		if ok, err := deps.HasUser(r.Context()); err != nil {
			http.Error(w, "Failed to load profile", http.StatusInternalServerError)
			return
		} else if !ok {
			http.Redirect(w, r, "/setup", http.StatusFound)
			return
		}
		next.ServeHTTP(w, r)
	})
}
