package public

import (
	"context"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/skip2/go-qrcode"
	"pin/internal/config"
	"pin/internal/domain"
	"pin/internal/features/domains"
	"pin/internal/features/profilepicture"
	featuresettings "pin/internal/features/settings"
)

type Dependencies interface {
	featuresettings.Store
	domains.Store
	domains.Protector
	profilepicture.Store
	Config() config.Config
	HasUser(ctx context.Context) (bool, error)
	GetOwnerUser(ctx context.Context) (domain.User, error)
	FindUserByIdentifier(ctx context.Context, identifier string) (domain.User, error)
	GetUserByPrivateToken(ctx context.Context, token string) (domain.User, error)
	Reserved() map[string]struct{}
	RenderTemplate(w http.ResponseWriter, name string, data interface{}) error
	BaseURL(r *http.Request) string
	GetSession(r *http.Request, name string) (*sessions.Session, error)
	EnsureCSRF(session *sessions.Session) string
	ValidateCSRF(session *sessions.Session, token string) bool
	CheckIdentifierCollisions(ctx context.Context, identifiers []string, excludeID int) error
	CreateUser(ctx context.Context, username, email, role, passwordHash, totpSecret, themeProfile, privateToken string) (int64, error)
	UpdatePrivateToken(ctx context.Context, userID int, token string) error
	UpsertUserIdentifiers(ctx context.Context, userID int, username string, aliases []string, email string) error
	AuditAttempt(ctx context.Context, actorID int, action, target string, meta map[string]string)
	AuditOutcome(ctx context.Context, actorID int, action, target string, err error, meta map[string]string)
}

// Handler serves public endpoints, including QR and profile pages.
type Handler struct {
	deps Dependencies
}

func NewHandler(deps Dependencies) Handler { return Handler{deps: deps} }

// QR renders a PNG QR code for the provided data string.
func (Handler) QR(w http.ResponseWriter, r *http.Request) {
	data := r.URL.Query().Get("data")
	if data == "" {
		http.Error(w, "Missing data", http.StatusBadRequest)
		return
	}
	if len(data) > 2048 {
		http.Error(w, "Data too long", http.StatusBadRequest)
		return
	}
	png, err := qrcode.Encode(data, qrcode.Medium, 256)
	if err != nil {
		http.Error(w, "Failed to generate QR", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "image/png")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(png)
}
