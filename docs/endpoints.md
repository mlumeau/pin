# Endpoints

This is a high-level overview of the main routes. See feature `routes.go` files for the full list.
For the normative PINC contract, see RFC-PINC.md.

## PINC endpoints (RFC-PINC)
All paths are relative to the Base URL.

### Owner identity
- `/json` - owner identity (public) Canonical JSON
- `/xml`, `/txt`, `/vcf` - owner identity alternate formats
- `/` - landing or profile page

### Public identities
- `/{handle}.json` - identity (public) Canonical JSON
- `/{handle}` - identity page or content negotiation
- `/{handle}.xml`, `/{handle}.txt`, `/{handle}.vcf` - alternate formats

### Private identities
- `/p/{...}.json` - identity (private) Canonical JSON
- `/p/{...}` - private page or content negotiation
- `/p/{...}.xml`, `/p/{...}.txt`, `/p/{...}.vcf` - alternate formats

### Profile pictures
- `/{handle}/profile-picture?s=160&format=webp` - public profile picture
- `/p/{...}/profile-picture?s=160&format=webp` - private profile picture

### Capability and schema
- `/.well-known/pinc` - capability document
- `/.well-known/pinc/identity` - JSON schema for canonical identity

## Public pages
- `/landing` - landing page regardless of default mode
- `/qr?data=...` - QR code PNG
- `/u/{username}` - shortcut to profile

## Setup and auth
- `/setup` - first-run setup (when no user exists)
- `/login` - login page
- `/logout` - logout
- `/invite/{token}` - invite flow

## Settings and admin
Most of these require a session.
- `/settings`, `/settings/profile`, `/settings/security`, `/settings/appearance`
- `/settings/profile/profile-picture/*` - select/delete/upload/alt
- `/settings/profile/verified-domains/*` - create/verify/delete
- `/settings/profile/social/bluesky` - connect Bluesky
- `/settings/admin/server`
- `/settings/admin/users` and `/settings/admin/users/{id}`
- `/settings/admin/invites/*`
- `/settings/admin/audit-log/download`

## Passkeys and OAuth
- `/passkeys/register/options`
- `/passkeys/register/finish`
- `/passkeys/delete`
- `/passkeys/login/options`
- `/passkeys/login/finish`
- `/oauth/github/start` and `/oauth/github/callback`
- `/oauth/reddit/start` and `/oauth/reddit/callback`

## Federation and other well-known
- `/.well-known/webfinger`
- `/.well-known/atproto-did`
- `/.well-known/pin-verify`
- `/users/{username}` - ActivityPub actor

## MCP
- `/mcp` - MCP JSON-RPC endpoint (when enabled)

## Health
- `/health/images` - image processing diagnostics
