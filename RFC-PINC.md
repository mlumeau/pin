# RFC — PINC: Personal Identity Node Contract

**Version**: pinc-1

**Status**: Draft  

**Author**: Maxime Lumeau  

**Date**: 2026-01-22

## License

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at:

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

## Abstract

This document specifies a minimal, interoperable contract for a Personal
Identity Node (PIN): a decentralized identity endpoint that provides a
single source of truth for one or more identities.

A conforming implementation exposes stable addressing for identities, a
required canonical JSON representation, optional alternate export formats,
and rules for distinguishing public and private identity views.  Private
views are accessed via unguessable (obfuscated) URLs and may include fields
that are not present in public views.

The contract supports a multi-identity instance with a designated Owner
Identity (primary identity) and optional additional identities.

## Status of This Document

This is a community working draft published by the PIN Project.  The
structure and normative language are inspired by common RFC conventions to
improve clarity and interoperability.
This document defines PINC version pinc-1.

## Table of Contents

1.  Introduction
2.  Conventions and Terminology
3.  Overview and Design Goals
4.  Multi-Identity Model and Owner Identity
5.  Public vs. Private Identity Views  
   5.1.  Identity Data Model and Custom Fields  
   5.2.  Per-Field Visibility  
   5.3.  Private View Addressing and Obfuscated URLs  
6.  Base URL, URL Structure, and Endpoints  
   6.1.  Base URL (Root, Subdomain, or Path Prefix)  
   6.2.  Owner Identity Endpoints  
   6.3.  Public Identity Endpoints  
   6.4.  Private Identity Endpoints  
   6.5.  Endpoint Recap  
7.  Representations  
   7.1.  Canonical JSON (Required)  
   7.2.  Alternate Export Formats (Optional)  
   7.3.  JSON Schema Endpoint (Optional)  
   7.4.  Content Negotiation (Optional)  
8.  Profile Pictures
9.  Capability Declaration (Optional)
10. HTTP Caching Semantics
11. Error Handling
12. Security Considerations
13. Privacy Considerations
14. References  
Appendix A.  Examples (Informative)  
Appendix B.  Example Storage Schema (Informative)

## 1. Introduction


Identity on the web is often fragmented across platforms, formats, and
protocols.  A Personal Identity Node (PIN) provides a canonical, long-lived
place to publish identity data with predictable semantics and stable URLs.
The Personal Identity Node Contract (PINC) defines the minimum interoperable
surface a PIN must expose.

PINC specifies:

- A Base URL and endpoint conventions for serving identity data.
- A required canonical JSON representation.
- Optional alternate export formats.
- A dual-view model (public vs. private), with per-field visibility
   evaluated by the server.
- An obfuscated URL mechanism for private views.

PINC does not define how identities are edited, authenticated, synchronized,
discovered, federated, or stored.

## 2. Conventions and Terminology


The key words "MUST", "MUST NOT", "REQUIRED", "SHALL", "SHALL NOT",
"SHOULD", "SHOULD NOT", "RECOMMENDED", "MAY", and "OPTIONAL" in this
document are to be interpreted as described in RFC 2119.

**PIN:**  
Personal Identity Node.  An implementation that serves identity data
under a Base URL as defined by this contract.

**PINC:**  
Personal Identity Node Contract.  The interoperability specification
defined by this document.

**Base URL:**  
The deployment base URL where the PIN is reachable.  The Base URL MAY
include a path prefix (i.e., it is not restricted to scheme/host/port).
All endpoint paths in this document are relative to the Base URL.

**Identity:**  
A subject that needs an identity representation.  An identity MAY
represent a person, organization, group, project, device, pet, or any
other entity that benefits from stable identification and contact/link
metadata.

**Handle:**  
The canonical identifier of an identity within a PIN (commonly a
username-like string).  For interoperable URL behavior, PINC RECOMMENDS
the handle profile `^[a-z0-9._-]+$` (lowercase ASCII letters, digits,
dot, underscore, hyphen).  PIN implementations SHOULD canonicalize
handles to lowercase and SHOULD reject characters outside this profile.

**Owner Identity:**  
The designated primary identity of a PIN instance.  The Owner Identity
provides required default endpoints (Section 6.2).

**Public View:**  
The publicly accessible view of an identity.  Public views contain only
the fields the server chooses to expose publicly.

**Private View:**  
A more complete view that may include fields that are not present in the
public view.  Private views are accessed via an unguessable (obfuscated)
URL (Section 5.3).

## 3. Overview and Design Goals


A conforming PIN is expected to be:

- Stable: URLs and semantics are long-lived.
- Deterministic: identical identity state yields identical representations.
- Decentralized: any party can host their own PIN at a chosen Base URL.
- Single source of truth: the Base URL is authoritative for the identity
   data it serves.
- Local-first: core functionality does not require third-party services.
- Consumer-friendly: humans and software can rely on predictable endpoints.

## 4. Multi-Identity Model and Owner Identity


A PIN instance MAY host multiple identities.

The PIN instance MUST designate exactly one identity as the Owner Identity.
The Owner Identity provides required default endpoints under the Base URL
(Section 6.2).  The Base URL path "/" itself MAY be used for a landing page,
setup flow, redirect, or other content.

Additional identities, if present, MUST be addressable through per-identity
endpoints (Section 6.3) and MAY have different exposure policies (public,
private, or both).

Identity enumeration is NOT required by PINC.  If a PIN provides any
enumeration endpoint, it SHOULD be disabled by default.

## 5. Public vs. Private Identity Views


PINC defines a dual-view model:

- Public View: intended for general publication.
- Private View: intended for selective sharing.

Both views refer to the same underlying identity record but differ in which
fields are included.  The practical rule for consumers is:

   If a field is present, it is intended for the consumer of that view.
   Fields that are not intended for the consumer MUST be omitted.

### 5.1. Identity Data Model and Custom Fields


Beyond the required core fields in Canonical JSON (Section 7.1.2), the
identity data model is implementation-defined.

Implementations MAY support user-defined custom fields.  Custom fields MAY
be structured (objects/arrays) or unstructured (strings).  If custom fields
exist, they SHOULD be exposed in Canonical JSON in a predictable location
(e.g., "custom_fields").  Custom fields MUST be subject to the same
public/private view rules as built-in fields.

PINC does not constrain which fields exist.

### 5.2. Per-Field Visibility


PINC assumes that field exposure is evaluated on a per-field basis.  For a
given identity record, each field is either exposed in the Public View or
withheld from the Public View and only exposed in the Private View (if a
private view is supported).

Implementations MUST enforce the following behavior:

- Public views MUST NOT include fields classified as private.
- Private views MAY include both public and private fields.

How a field becomes public or private is implementation-defined.  Two common
models are supported by this contract:

- Server policy model: The server defines a fixed or constrained set of
   rules that determine visibility (e.g., "emails are always private",
   "display_name is always public").  The server MAY restrict which fields
   can be hidden or forced public.

- User-controlled model: The identity owner chooses, on a per-field basis,
   which fields are public and which are private, potentially within bounds
   set by the server (e.g., the server may require some minimal public
   fields such as display_name).

PINC does not mandate a specific policy mechanism or data structure for
storing visibility rules.  If a PIN supports user-controlled visibility, it
SHOULD provide a predictable way to manage those settings, but the settings
themselves SHOULD NOT be included in public or private representations.

Default behavior for unspecified fields is implementation-defined.  PINC
RECOMMENDS default-public, but a PIN MAY choose more conservative defaults.

### 5.3. Private View Addressing and Obfuscated URLs


Private Views are accessed via obfuscated URLs that MUST be unguessable.
Possession of the full private URL is treated as the authorization mechanism
unless the PIN adds stronger access control.

Requirements:

- The private URL MUST contain at least one component with at least 128 bits
   of entropy.
- The high-entropy component MUST be generated using a cryptographically
   secure RNG.
- No combination of publicly known inputs (handle, email, etc.) may be used
   to derive the entire private URL.
- The private URL MUST be safe for use in URLs (e.g., base64url or hex).
- PINs MUST return 404 for invalid or unknown private URLs.
- PINs SHOULD avoid any endpoint that allows private URL enumeration.

A private URL MAY be composed of multiple path segments.  Some segments MAY
be derived (e.g., a short hash of an internal identity identifier), but the
full URL MUST remain unguessable due to the high-entropy segment.

Examples (informative):
   `/p/{token}`
   `/p/{derived}/{token}`
   `/private/{derived}/{token}`

## 6. Base URL, URL Structure, and Endpoints


### 6.1. Base URL (Root, Subdomain, or Path Prefix)


A PIN is deployed at a Base URL.  The Base URL MAY be:

- a domain root (https://example.com),
- a dedicated subdomain (https://pin.example.com), or
- a path prefix under a domain (https://example.com/pin).

Endpoint paths in this specification are relative to the Base URL.

### 6.2. Owner Identity Endpoints


REQUIRED:

- `GET {BaseURL}/json`
   Returns Canonical JSON for the Owner Identity (Public View).

OPTIONAL:

- `GET {BaseURL}/`
   MAY return a landing page, setup page, redirect, an Owner profile page,
   or any other content.

- `GET {BaseURL}/xml`, `/txt`, `/vcf`
   Optional alternate formats for the Owner Identity if supported.

### 6.3. Public Identity Endpoints


REQUIRED:

- `GET {BaseURL}/{handle}.json`
   MUST return Canonical JSON for the Public View of the identity.

OPTIONAL:

- `GET {BaseURL}/{handle}`
   MAY return an HTML profile page for the Public View, or MAY return a
   negotiated representation (Section 7.4).

- `GET {BaseURL}/{handle}.xml`, `.txt`, `.vcf`
   Optional alternate formats for the identity if supported.

### 6.4. Private Identity Endpoints


A PIN MAY provide private endpoints for identities.

If provided, private endpoints MUST be addressed by obfuscated URLs (Section
5.3).  The private JSON endpoint is REQUIRED if private views are supported:

- `GET {BaseURL}/p/{...}.json`
   MUST return Canonical JSON for the Private View of the identity.

OPTIONAL:

- `GET {BaseURL}/p/{...}`
   MAY return an HTML page for the Private View, or MAY return a negotiated
   representation (Section 7.4).

- `GET {BaseURL}/p/{...}.xml`, `.txt`, `.vcf`
   Optional alternate formats for the private view if supported.

### 6.5. Endpoint Recap


The following table summarizes endpoint requirements.  All paths are
relative to the Base URL.

| Endpoint                    | Meaning                      | Status   |
|-----------------------------|------------------------------|----------|
| `/json`             | **Owner identity (public) JSON** | **REQUIRED** |
| `/{handle}.json`    | **Identity (public) JSON**       | **REQUIRED** |
| `/`                         | Landing/setup/other content  | OPTIONAL |
| `/xml`, `/txt`, `/vcf`      | Owner alt formats            | OPTIONAL |
| `/{handle}`                 | Identity page or negotiation | OPTIONAL |
| `/{handle}.xml/.txt/.vcf`   | Identity alt formats         | OPTIONAL |
| `/p/{...}.json`             | Identity (private) JSON      | OPTIONAL |
| `/p/{...}`                  | Private page or negotiation  | OPTIONAL |
| `/p/{...}.xml/.txt/.vcf`    | Private alt formats          | OPTIONAL |
| `/{handle}/profile-picture` | Public profile picture       | OPTIONAL |
| `/p/{...}/profile-picture`  | Private profile picture      | OPTIONAL |
| `/.well-known/pinc`         | Capability document          | OPTIONAL |
| `/.well-known/pinc/identity`| JSON Schema for identity     | OPTIONAL |

## 7. Representations


### 7.1 Canonical JSON (Required)


#### 7.1.1 General


Canonical JSON is REQUIRED and MUST be supported for the Owner Identity and
for non-owner identities (Public View).  If private views are supported,
Canonical JSON MUST also be supported for Private Views.

Canonical JSON MUST be deterministic for a given identity state and view.

#### 7.1.2 Envelope and Required Fields


The response MUST be a JSON object containing:
```json
{
   "meta": { ... },
   "identity": { ... }
}
```

Where:
```
meta (object):  
   version  (string): schema version identifier (e.g., "pinc-1")
   base_url (string): the Base URL of the PIN
   view     (string): "public" or "private"
   subject  (string): a stable opaque identifier for the identity
   rev      (string): a deterministic revision identifier for this view
   self     (string): OPTIONAL.  Absolute URL of this representation.

identity (object):
   handle       (string): canonical handle within this PIN
   display_name (string): human-facing name
   url          (string): canonical public profile URL (HTML or JSON URL)
   updated_at   (string): RFC 3339 timestamp of last semantic update
   ... additional fields as defined by the implementation ...
```

Notes:

- "updated_at" and "rev" serve different purposes.  "updated_at" is a
   human-meaningful timestamp; "rev" is a deterministic fingerprint for
   caching, deduplication, and integrity.  A PIN MAY compute rev as a hash
   over the canonicalized identity object for the given view.
- "self" is not the same as "base_url".  "base_url" identifies the
   deployment prefix; "self" identifies the specific representation URL.
- The server's internal visibility policy MUST NOT be included in this
   representation (Section 5.2).

PINs MAY include additional fields in identity, including custom fields.
PINs MUST omit non-public fields from Public Views, and MAY include them in
Private Views.

### 7.2 Alternate Export Formats (Optional)


PINs MAY support additional formats (XML, TXT, vCard, others).

If supported, PINs MUST clearly document which export formats are available
and how to request them.  A capability document (Section 9) is one possible
mechanism.

### 7.3 JSON Schema Endpoint (Optional)


PINs SHOULD publish a JSON Schema for the canonical identity representation
to facilitate integration and reimplementation.

If provided, it SHOULD be available at:

- `GET {BaseURL}/.well-known/pinc/identity`

### 7.4 Content Negotiation (Optional)


PINs MAY support content negotiation using the Accept header.

If supported:

- `GET {BaseURL}/{handle}` with `Accept: application/json` MAY return Canonical
   JSON.
- `GET {BaseURL}/p/{...}` with `Accept: application/json` MAY return Canonical
   JSON.

## 8. Profile Pictures


PINs MAY provide a profile picture resource.

If supported, the endpoint SHOULD allow selection of size and output format
for portability and caching.

Public View example endpoint:

   `GET {BaseURL}/{handle}/profile-picture?s=160&format=webp`

Private View example endpoint (if private views are supported):

   `GET {BaseURL}/p/{...}/profile-picture?s=160&format=webp`

Implementations SHOULD validate inputs and clamp size parameters to safe
bounds.  Profile picture format support is implementation-defined.

## 9. Capability Declaration (Optional)


A PIN implementation MAY expose a capability document to let clients
discover which parts of PINC are implemented and which representations are 
available.

If present, the capability document MUST be served at:

   ``GET {BaseURL}/.well-known/pinc``

The capability document SHOULD be JSON with media type application/json.
It SHOULD be stable and cacheable, and it SHOULD only describe
client-visible behavior (not internal configuration).

The capability document SHOULD include:

   - `pinc_version`: the PINC version implemented (e.g., "pinc-1")
   - `base_url`: the Base URL this document applies to
   - `export_formats`: list of supported export formats (MUST include 
      "json" when PINC is supported, e.g., ["json","txt","csv","xml"])
   - `views`: which views are available ("public", optionally "private")
   - `media_formats`: supported profile-picture output formats (e.g.,
      ["webp","png","jpeg","gif"])

Example (informative):
```json
   {
      "pinc_version": "pinc-1",
      "base_url": "https://example.com/pin",
      "export_formats": ["json","txt","xml"],
      "views": ["public","private"],
      "media_formats": ["webp","png","jpeg","gif"]
   }
```

## 10. HTTP Caching Semantics


Public Views are intended for caching.

Private Views carry higher privacy risk; private JSON endpoints SHOULD send
"Cache-Control: private, no-store" unless explicitly configured otherwise.

## 11. Error Handling


PINs MUST use standard HTTP status codes (404 for unknown handle/token, 400
for invalid parameters, etc.).

## 12. Security Considerations


Deployments MUST use HTTPS.  Private URLs act as bearer secrets and MUST be
high-entropy.

## 13. Privacy Considerations


PINs SHOULD treat private URLs as secrets and SHOULD avoid exposing internal
policy or hidden-field structure through public or private views.

## 14. References


[RFC2119]  Bradner, S., "Key words for use in RFCs to Indicate Requirement
            Levels", RFC 2119, March 1997.

[RFC3339]  Klyne, G. and C. Newman, "Date and Time on the Internet:
            Timestamps", RFC 3339, July 2002.

## Appendix A.  Examples (Informative)


### A.1.  Example Canonical JSON (Public View)
```json
{
   "meta": {
      "version": "pinc-1",
      "base_url": "https://example.com/pin",
      "self": "https://example.com/pin/alice.json",
      "view": "public",
      "subject": "urn:pin:subject:Q2hhbmdlTWU",
      "rev": "sha256:1b2c...9f"
   },
   "identity": {
      "handle": "alice",
      "display_name": "Alice Example",
      "url": "https://example.com/pin/alice",
      "updated_at": "2026-01-22T08:00:00Z",
      "summary": "Systems engineer.",
      "custom_fields": {
      "favorite_color": "blue"
      }
   }
}
```

### A.2.  Example Canonical JSON (Private View)
```json
{
   "meta": {
      "version": "pinc-1",
      "base_url": "https://example.com",
      "self": "https://example.com/p/1a2b3c/7f3e5d...c2.json",
      "view": "private",
      "subject": "urn:pin:subject:Q2hhbmdlTWU",
      "rev": "sha256:aa11...ff"
   },
   "identity": {
      "handle": "alice",
      "display_name": "Alice Example",
      "url": "https://example.com/alice",
      "updated_at": "2026-01-22T08:00:00Z",
      "summary": "Systems engineer.",
      "emails": [
      { "value": "alice@example.com", "verified": true }
      ],
      "custom_fields": {
      "tax_id": "FRXX..."
      }
   }
}
```

## Appendix B.  Example Storage Schema (Informative)


This appendix provides a minimal example of how a PIN implementation might
store identities.  Storage is out of scope for PINC; this is non-normative.
```sql
CREATE TABLE identities (
   id               INTEGER PRIMARY KEY,
   handle           TEXT NOT NULL UNIQUE,
   is_owner         INTEGER NOT NULL DEFAULT 0,
   display_name     TEXT NOT NULL,
   summary          TEXT NOT NULL DEFAULT '',
   visibilities     TEXT NOT NULL,   -- user-defined visibility map object
   updated_at       TEXT NOT NULL    -- RFC 3339
);
```
