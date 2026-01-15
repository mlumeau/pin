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
	UpdatePrivateToken(ctx context.Context, userID int, token string) error
	BaseURL(r *http.Request) string
	ResetAllUserThemes(ctx context.Context, themeValue string) error
	ListUsersPaged(ctx context.Context, query, sort, dir string, limit, offset int) ([]domain.User, int, error)
	ListUsers(ctx context.Context) ([]domain.User, error)
	ListInvites(ctx context.Context) ([]domain.Invite, error)
	ListAuditLogs(ctx context.Context, limit, offset int) ([]domain.AuditLog, error)
	CountAuditLogs(ctx context.Context) (int, error)
	ListAllAuditLogs(ctx context.Context) ([]domain.AuditLog, error)
	GetUserByID(ctx context.Context, id int) (domain.User, error)
	DeleteUser(ctx context.Context, userID int) error
	UpdateUser(ctx context.Context, user domain.User) error
	UpsertUserIdentifiers(ctx context.Context, userID int, username string, aliases []string, email string) error
	CheckIdentifierCollisions(ctx context.Context, identifiers []string, excludeID int) error
	Reserved() map[string]struct{}
	ListPasskeys(ctx context.Context, userID int) ([]domain.Passkey, error)
	ListProfilePictures(ctx context.Context, userID int) ([]domain.ProfilePicture, error)
	CreateProfilePicture(ctx context.Context, userID int, filename, alt string) (int64, error)
	ListDomainVerifications(ctx context.Context, userID int) ([]domain.DomainVerification, error)
	UpsertDomainVerification(ctx context.Context, userID int, domainName, token string) error
	DeleteDomainVerification(ctx context.Context, userID int, domainName string) error
	MarkDomainVerified(ctx context.Context, userID int, domainName string) error
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
