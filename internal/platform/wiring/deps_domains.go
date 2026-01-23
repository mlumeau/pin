package wiring

import (
	"context"

	"pin/internal/domain"
)

// Domains.
func (d Deps) ListDomainVerifications(ctx context.Context, identityID int) ([]domain.DomainVerification, error) {
	return d.repos.Domains.ListDomainVerifications(ctx, identityID)
}

func (d Deps) UpsertDomainVerification(ctx context.Context, identityID int, domainName, token string) error {
	return d.repos.Domains.UpsertDomainVerification(ctx, identityID, domainName, token)
}

func (d Deps) DeleteDomainVerification(ctx context.Context, identityID int, domainName string) error {
	return d.repos.Domains.DeleteDomainVerification(ctx, identityID, domainName)
}

func (d Deps) MarkDomainVerified(ctx context.Context, identityID int, domainName string) error {
	return d.repos.Domains.MarkDomainVerified(ctx, identityID, domainName)
}

func (d Deps) HasDomainVerification(ctx context.Context, identityID int, domainName string) (bool, error) {
	return d.repos.Domains.HasDomainVerification(ctx, identityID, domainName)
}

func (d Deps) ProtectedDomain(ctx context.Context) string {
	return d.repos.Domains.ProtectedDomain(ctx)
}

func (d Deps) SetProtectedDomain(ctx context.Context, domainName string) error {
	return d.repos.Domains.SetProtectedDomain(ctx, domainName)
}
