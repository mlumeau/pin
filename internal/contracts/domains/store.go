package domains

import (
	"context"

	"pin/internal/domain"
)

// Repository defines persistence operations for verified domains.
type Repository interface {
	ListDomainVerifications(ctx context.Context, userID int) ([]domain.DomainVerification, error)
	ListVerifiedDomains(ctx context.Context, userID int) ([]string, error)
	UpsertDomainVerification(ctx context.Context, userID int, domain, token string) error
	DeleteDomainVerification(ctx context.Context, userID int, domain string) error
	MarkDomainVerified(ctx context.Context, userID int, domain string) error
	UpdateUserVerifiedDomains(ctx context.Context, userID int, domains []string) error
	HasDomainVerification(ctx context.Context, userID int, domain string) (bool, error)
	ProtectedDomain(ctx context.Context) string
	SetProtectedDomain(ctx context.Context, domain string) error
}
