package wiring

import (
	"context"

	"github.com/go-webauthn/webauthn/webauthn"
	"pin/internal/domain"
)

// Passkeys.
func (d Deps) ListPasskeys(ctx context.Context, userID int) ([]domain.Passkey, error) {
	return d.repos.Passkeys.ListPasskeys(ctx, userID)
}

func (d Deps) LoadPasskeyCredentials(ctx context.Context, userID int) ([]webauthn.Credential, error) {
	return d.repos.Passkeys.LoadPasskeyCredentials(ctx, userID)
}

func (d Deps) InsertPasskey(ctx context.Context, userID int, name string, credential webauthn.Credential) error {
	return d.repos.Passkeys.InsertPasskey(ctx, userID, name, credential)
}

func (d Deps) UpdatePasskeyCredential(ctx context.Context, userID int, credentialID string, credential webauthn.Credential) error {
	return d.repos.Passkeys.UpdatePasskeyCredential(ctx, userID, credentialID, credential)
}

func (d Deps) DeletePasskey(ctx context.Context, userID, id int) error {
	return d.repos.Passkeys.DeletePasskey(ctx, userID, id)
}
