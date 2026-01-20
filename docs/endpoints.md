# Endpoints

This is a high-level overview of the main routes. See feature `routes.go` files for the full list.

## Public pages
- `/` - landing page (or profile, depending on settings)
- `/landing` - landing page regardless of default mode
- `/qr?data=...` - QR code PNG
- `/{username}` - public profile page
- `/u/{username}` - shortcut to profile
- `/p/{hash}/{token}` - private profile view
- `/profile-picture/{username}?s=160&format=webp` - profile picture (size/format)

## Identity exports
- `/json`, `/xml`, `/txt`, `/vcf` - owner identity exports
- `/{identifier}.json|xml|txt|vcf` - identity exports by username/alias/email
- `/.well-known/identity.json|xml|txt|vcf` - standard identity endpoints
- `/identity.schema.json` - JSON schema for identity export

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

## Federation and well-known
- `/.well-known/webfinger`
- `/.well-known/atproto-did`
- `/.well-known/pin-verify`
- `/users/{username}` - ActivityPub actor

## MCP
- `/mcp` - MCP JSON-RPC endpoint (when enabled)

## Health
- `/health/images` - image processing diagnostics
