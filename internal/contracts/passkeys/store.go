package passkeys

import (
	"context"

	"github.com/go-webauthn/webauthn/webauthn"
	"pin/internal/domain"
)

// Repository defines persistence operations for passkeys.
type Repository interface {
	ListPasskeys(ctx context.Context, userID int) ([]domain.Passkey, error)
	LoadPasskeyCredentials(ctx context.Context, userID int) ([]webauthn.Credential, error)
	InsertPasskey(ctx context.Context, userID int, name string, credential webauthn.Credential) error
	UpdatePasskeyCredential(ctx context.Context, userID int, credentialID string, credential webauthn.Credential) error
	DeletePasskey(ctx context.Context, userID, id int) error
}
