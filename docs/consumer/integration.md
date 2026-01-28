# Consumer integration

This document describes how to integrate with a PIN server as a consumer.
For the normative interoperability contract, see [RFC-PINC.md](../../RFC-PINC.md).

## Scope

PINC is intentionally minimal. It standardizes read-only identity exports and
how to address them. It does not define discovery, enumeration, auth, write
APIs, or federation. Consumers should treat anything outside PINC as
implementation-specific or out of scope.

## What a consumer can rely on

PINC defines a Base URL and a set of stable endpoints that expose a canonical
JSON representation of an identity. Required endpoints:

- `{BaseURL}/json` for the Owner identity (public view)
- `{BaseURL}/{handle}.json` for a public identity

If a server supports private views, it MUST also expose:

- `{BaseURL}/p/{...}.json` for the private view

The canonical JSON envelope and required fields are defined in RFC-PINC.

## Capability discovery (optional)

If present, the capability document at `/.well-known/pinc` advertises:

- the PINC version implemented
- supported export formats (json, txt, xml, vcf, etc.)
- available views (public, optionally private)
- supported profile picture media formats

Use this to feature-detect optional behaviors and formats instead of guessing.

## Content negotiation (optional)

A PIN MAY support content negotiation. If it does, requesting
`Accept: application/json` from `/{handle}` or `/p/{...}` may return canonical
JSON. If you need JSON, the `.json` endpoints are the most reliable.

## Caching guidance

Public views are intended to be cached. The `meta.rev` field is a deterministic
revision identifier that can be used to detect semantic changes even if the
HTTP cache headers are not present. Private views should generally be treated
as non-cacheable unless the server explicitly indicates otherwise.

## Error handling expectations

PINC relies on standard HTTP status codes. Consumers should handle:

- `404` for unknown handles or private tokens
- `400` for invalid parameters
- `401` or `403` if a server adds stronger access control

## Privacy and security

Private view URLs are bearer secrets. Treat them like credentials and do not
log or share them casually. Only request private endpoints when you intend to
consume private data.

## Integration checklist

- Fetch `{BaseURL}/json` for the Owner identity if you need a default profile.
- Use `{BaseURL}/{handle}.json` for public identity data.
- Prefer `.json` endpoints over negotiation unless explicitly supported.
- Check `/.well-known/pinc` to discover optional formats and views.
- Treat `meta.rev` as a change detector and `identity.updated_at` as a
  human-meaningful update timestamp.
- Avoid relying on enumeration or discovery unless the implementation
  documents it explicitly.

## Out of scope

If you need any of the following, expect implementation-specific behavior or
additional documents:

- identity discovery or enumeration
- authenticated access to private data beyond bearer URLs
- write/update APIs
- synchronization or federation
- webhooks or real-time updates
