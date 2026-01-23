package admin

import (
	"context"
	"net/http"

	"github.com/gorilla/sessions"
	"pin/internal/config"
	"pin/internal/domain"
	featuresettings "pin/internal/features/settings"
)

type Dependencies interface {
	featuresettings.Store
	Config() config.Config
	GetSession(r *http.Request, name string) (*sessions.Session, error)
	EnsureCSRF(session *sessions.Session) string
	ValidateCSRF(session *sessions.Session, token string) bool
	CurrentUser(r *http.Request) (domain.User, error)
	CurrentIdentity(r *http.Request) (domain.Identity, error)
	UpdateIdentityPrivateToken(ctx context.Context, identityID int, token string) error
	BaseURL(r *http.Request) string
	ResetAllUserThemes(ctx context.Context, themeValue string) error
	ListIdentitiesPaged(ctx context.Context, query, sort, dir string, limit, offset int) ([]domain.Identity, int, error)
	ListIdentities(ctx context.Context) ([]domain.Identity, error)
	ListInvites(ctx context.Context) ([]domain.Invite, error)
	ListAuditLogs(ctx context.Context, limit, offset int) ([]domain.AuditLog, error)
	CountAuditLogs(ctx context.Context) (int, error)
	ListAllAuditLogs(ctx context.Context) ([]domain.AuditLog, error)
	GetUserByID(ctx context.Context, id int) (domain.User, error)
	GetIdentityByUserID(ctx context.Context, userID int) (domain.Identity, error)
	DeleteUser(ctx context.Context, userID int) error
	UpdateUser(ctx context.Context, user domain.User) error
	UpdateIdentity(ctx context.Context, identity domain.Identity) error
	CheckHandleCollision(ctx context.Context, handle string, excludeID int) error
	Reserved() map[string]struct{}
	ListPasskeys(ctx context.Context, userID int) ([]domain.Passkey, error)
	ListProfilePictures(ctx context.Context, identityID int) ([]domain.ProfilePicture, error)
	CreateProfilePicture(ctx context.Context, identityID int, filename, alt string) (int64, error)
	ListDomainVerifications(ctx context.Context, identityID int) ([]domain.DomainVerification, error)
	UpsertDomainVerification(ctx context.Context, identityID int, domainName, token string) error
	DeleteDomainVerification(ctx context.Context, identityID int, domainName string) error
	MarkDomainVerified(ctx context.Context, identityID int, domainName string) error
	ProtectedDomain(ctx context.Context) string
	RenderTemplate(w http.ResponseWriter, name string, data interface{}) error
	AuditAttempt(ctx context.Context, actorID int, action, target string, meta map[string]string)
	AuditOutcome(ctx context.Context, actorID int, action, target string, err error, meta map[string]string)
}

type Handler struct {
	deps Dependencies
}

func NewHandler(deps Dependencies) Handler {
	return Handler{deps: deps}
}
