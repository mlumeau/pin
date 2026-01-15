package contracts

import (
	"pin/internal/contracts/audit"
	"pin/internal/contracts/domains"
	"pin/internal/contracts/identifiers"
	"pin/internal/contracts/invites"
	"pin/internal/contracts/passkeys"
	"pin/internal/contracts/profilepictures"
	"pin/internal/contracts/settings"
	"pin/internal/contracts/users"
)

// Repos groups feature-specific repositories for injection into services and handlers.
type Repos struct {
	Users           users.Repository
	Invites         invites.Repository
	Passkeys        passkeys.Repository
	Audit           audit.Repository
	Domains         domains.Repository
	Identifiers     identifiers.Repository
	ProfilePictures profilepictures.Repository
	Settings        settings.Repository
}
