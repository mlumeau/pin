package wiring

import (
	"context"

	"pin/internal/domain"
)

// Domains.
func (d Deps) ListDomainVerifications(ctx context.Context, userID int) ([]domain.DomainVerification, error) {
	return d.repos.Domains.ListDomainVerifications(ctx, userID)
}

func (d Deps) UpsertDomainVerification(ctx context.Context, userID int, domainName, token string) error {
	return d.repos.Domains.UpsertDomainVerification(ctx, userID, domainName, token)
}

func (d Deps) DeleteDomainVerification(ctx context.Context, userID int, domainName string) error {
	return d.repos.Domains.DeleteDomainVerification(ctx, userID, domainName)
}

func (d Deps) MarkDomainVerified(ctx context.Context, userID int, domainName string) error {
	return d.repos.Domains.MarkDomainVerified(ctx, userID, domainName)
}

func (d Deps) HasDomainVerification(ctx context.Context, userID int, domainName string) (bool, error) {
	return d.repos.Domains.HasDomainVerification(ctx, userID, domainName)
}

func (d Deps) ProtectedDomain(ctx context.Context) string {
	return d.repos.Domains.ProtectedDomain(ctx)
}

func (d Deps) SetProtectedDomain(ctx context.Context, domainName string) error {
	return d.repos.Domains.SetProtectedDomain(ctx, domainName)
}
