package domains

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"

	domainpkg "pin/internal/domain"
	"pin/internal/features/identity"
)

type Store interface {
	ListDomainVerifications(ctx context.Context, identityID int) ([]domainpkg.DomainVerification, error)
	UpsertDomainVerification(ctx context.Context, identityID int, domain, token string) error
	DeleteDomainVerification(ctx context.Context, identityID int, domain string) error
	MarkDomainVerified(ctx context.Context, identityID int, domain string) error
}

type Protector interface {
	HasDomainVerification(ctx context.Context, identityID int, domain string) (bool, error)
	ProtectedDomain(ctx context.Context) string
	SetProtectedDomain(ctx context.Context, domain string) error
}

type Service struct {
	store Store
}

func NewService(store Store) Service {
	return Service{store: store}
}

func (s Service) VerifyDomain(ctx context.Context, identityID int, domain string) ([]domainpkg.DomainVerification, []string, error) {
	rows, err := s.store.ListDomainVerifications(ctx, identityID)
	if err != nil {
		return nil, nil, err
	}
	token := ""
	for _, row := range rows {
		if row.Domain == domain {
			token = row.Token
			break
		}
	}
	if token == "" {
		return nil, nil, errors.New("domain not found")
	}
	ok, err := checkDomainVerification(ctx, domain, token)
	if err != nil {
		return nil, nil, err
	}
	if !ok {
		return nil, nil, errors.New("token not found")
	}
	if err := s.store.MarkDomainVerified(ctx, identityID, domain); err != nil {
		return nil, nil, err
	}
	rows, err = s.store.ListDomainVerifications(ctx, identityID)
	if err != nil {
		return nil, nil, err
	}
	verified := identity.VerifiedDomains(rows)
	found := false
	for _, item := range verified {
		if item == domain {
			found = true
			break
		}
	}
	if !found {
		verified = append(verified, domain)
	}
	return rows, verified, nil
}

func (s Service) DeleteDomain(ctx context.Context, identityID int, domain string) ([]domainpkg.DomainVerification, []string, error) {
	if err := s.store.DeleteDomainVerification(ctx, identityID, domain); err != nil {
		return nil, nil, err
	}
	rows, err := s.store.ListDomainVerifications(ctx, identityID)
	if err != nil {
		return nil, nil, err
	}
	verified := identity.VerifiedDomains(rows)
	remaining := make([]string, 0, len(verified))
	for _, item := range verified {
		if !strings.EqualFold(item, domain) {
			remaining = append(remaining, item)
		}
	}
	return rows, remaining, nil
}

func (s Service) CreateDomains(ctx context.Context, identityID int, domains []string, tokenFactory func() string) ([]domainpkg.DomainVerification, []string, error) {
	rows, err := s.syncDomainVerifications(ctx, identityID, domains, tokenFactory)
	if err != nil {
		return nil, nil, err
	}
	verified := identity.VerifiedDomains(rows)
	return rows, verified, nil
}

func (s Service) SeedDomains(ctx context.Context, identityID int, domains []string, tokenFactory func() string) []domainpkg.DomainVerification {
	rows := []domainpkg.DomainVerification{}
	for _, domain := range domains {
		token := tokenFactory()
		if err := s.store.UpsertDomainVerification(ctx, identityID, domain, token); err == nil {
			rows = append(rows, domainpkg.DomainVerification{IdentityID: identityID, Domain: domain, Token: token})
		}
	}
	return rows
}

func (s Service) syncDomainVerifications(ctx context.Context, identityID int, domains []string, tokenFactory func() string) ([]domainpkg.DomainVerification, error) {
	existing, err := s.store.ListDomainVerifications(ctx, identityID)
	if err != nil {
		return nil, err
	}
	seen := map[string]domainpkg.DomainVerification{}
	for _, row := range existing {
		seen[strings.ToLower(row.Domain)] = row
	}
	var rows []domainpkg.DomainVerification
	for _, domain := range domains {
		normalized := strings.ToLower(strings.TrimSpace(domain))
		if normalized == "" {
			continue
		}
		if row, ok := seen[normalized]; ok {
			rows = append(rows, row)
			continue
		}
		token := tokenFactory()
		if err := s.store.UpsertDomainVerification(ctx, identityID, normalized, token); err != nil {
			return nil, err
		}
		rows = append(rows, domainpkg.DomainVerification{IdentityID: identityID, Domain: normalized, Token: token})
	}
	toDelete := map[int]domainpkg.DomainVerification{}
	for _, row := range existing {
		toDelete[row.ID] = row
	}
	for _, row := range rows {
		delete(toDelete, row.ID)
	}
	for _, row := range toDelete {
		_ = s.store.DeleteDomainVerification(ctx, identityID, row.Domain)
	}
	return rows, nil
}

func (s Service) EnsureServerDomainVerification(ctx context.Context, baseURL string, identityID int, protector Protector, tokenFactory func() string) {
	if strings.TrimSpace(baseURL) == "" {
		return
	}
	baseDomain := normalizeDomain(strings.ToLower(strings.TrimSpace(baseURL)))
	if baseDomain == "" {
		return
	}
	if protected := protector.ProtectedDomain(ctx); protected != "" {
		return
	}
	if exists, err := protector.HasDomainVerification(ctx, identityID, baseDomain); err == nil && !exists {
		_ = protector.SetProtectedDomain(ctx, baseDomain)
		_, _ = s.syncDomainVerifications(ctx, identityID, []string{baseDomain}, tokenFactory)
	}
}

func checkDomainVerification(ctx context.Context, domain, token string) (bool, error) {
	domain = normalizeDomain(domain)
	if domain == "" || strings.TrimSpace(token) == "" {
		return false, nil
	}
	for _, scheme := range []string{"https", "http"} {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, scheme+"://"+domain+"/.well-known/pin-verify", nil)
		if err != nil {
			continue
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			continue
		}
		body, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			for _, line := range strings.Split(string(body), "\n") {
				if strings.TrimSpace(line) == token {
					return true, nil
				}
			}
		}
	}
	return false, nil
}
