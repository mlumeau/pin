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
	GetUserByID(ctx context.Context, id int) (domain.User, error)
	GetOwnerIdentity(ctx context.Context) (domain.Identity, error)
	GetIdentityByHandle(ctx context.Context, handle string) (domain.Identity, error)
	GetIdentityByPrivateToken(ctx context.Context, token string) (domain.Identity, error)
	Reserved() map[string]struct{}
	RenderTemplate(w http.ResponseWriter, name string, data interface{}) error
	BaseURL(r *http.Request) string
	GetSession(r *http.Request, name string) (*sessions.Session, error)
	CurrentUser(r *http.Request) (domain.User, error)
	CurrentIdentity(r *http.Request) (domain.Identity, error)
	EnsureCSRF(session *sessions.Session) string
	ValidateCSRF(session *sessions.Session, token string) bool
	CheckHandleCollision(ctx context.Context, handle string, excludeID int) error
	CreateUser(ctx context.Context, role, passwordHash, totpSecret, themeProfile string) (int64, error)
	CreateIdentity(ctx context.Context, identity domain.Identity) (int64, error)
	UpdateIdentityPrivateToken(ctx context.Context, identityID int, token string) error
	DeleteUser(ctx context.Context, userID int) error
	AuditAttempt(ctx context.Context, actorID int, action, target string, meta map[string]string)
	AuditOutcome(ctx context.Context, actorID int, action, target string, err error, meta map[string]string)
}

// Handler serves public endpoints.
type Handler struct {
	deps Dependencies
}

// NewHandler constructs a new handler.
func NewHandler(deps Dependencies) Handler { return Handler{deps: deps} }
