package public

import (
	"context"
	"net/http"

	"github.com/gorilla/sessions"
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

// Handler serves public endpoints.
type Handler struct {
	deps Dependencies
}

func NewHandler(deps Dependencies) Handler { return Handler{deps: deps} }
