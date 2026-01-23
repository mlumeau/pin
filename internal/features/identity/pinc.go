package identity

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"

	"pin/internal/domain"
)

// PincVersion is the current PINC version implemented by this server.
const PincVersion = "pinc-1"

// SubjectForIdentity returns a stable opaque subject identifier for an identity.
func SubjectForIdentity(identity domain.Identity) string {
	input := fmt.Sprintf("pin:%d:%s", identity.ID, strings.ToLower(strings.TrimSpace(identity.Handle)))
	sum := sha256.Sum256([]byte(input))
	return "urn:pin:subject:" + base64.RawURLEncoding.EncodeToString(sum[:])
}
