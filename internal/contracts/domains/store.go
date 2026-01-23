package domains

import (
	"context"

	"pin/internal/domain"
)

// Repository defines persistence operations for verified domains.
type Repository interface {
	ListDomainVerifications(ctx context.Context, identityID int) ([]domain.DomainVerification, error)
	ListVerifiedDomains(ctx context.Context, identityID int) ([]string, error)
	UpsertDomainVerification(ctx context.Context, identityID int, domain, token string) error
	DeleteDomainVerification(ctx context.Context, identityID int, domain string) error
	MarkDomainVerified(ctx context.Context, identityID int, domain string) error
	UpdateIdentityVerifiedDomains(ctx context.Context, identityID int, domains []string) error
	HasDomainVerification(ctx context.Context, identityID int, domain string) (bool, error)
	ProtectedDomain(ctx context.Context) string
	SetProtectedDomain(ctx context.Context, domain string) error
}
