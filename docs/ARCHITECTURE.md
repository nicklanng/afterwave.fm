# Architecture

High-level technical shape: surfaces, stack, auth, storage, providers, and global design. Deployment is in [Deployment](./DEPLOYMENT.md). Go style is in [Go style guide](./GO_STYLE.md).

---

## Surfaces and domains

- **www.afterwave.fm** — Landing and promo site when not logged in; when logged in it becomes the discovery and content site (browse, search, feed, account). Same frontend; behaviour depends on auth state.
- **handle.afterwave.fm** (e.g. barenakedapology.afterwave.fm) — Artist page. The **frontend** reads the subdomain (handle), calls the API for that artist’s data (e.g. `GET /v1/artists/:handle`), and renders the page. The **API does not care about subdomains**; routing by handle is a purely frontend concern. All artist/band **IDs are lowercase with no special characters**; artists can supply a **stylised name** for display (e.g. handle `barenakedapology`, display name “Bare Naked Apology”).
- **api.afterwave.fm** — API only. All API traffic goes here. **Versioned:** e.g. `GET api.afterwave.fm/v1/users/me`. The API never interprets Host for artist routing; the frontend decides what to fetch based on the current domain (www vs handle).

We run a **monolith**: one Go API, **two frontend codebases** — the **website** (www + artist pages) and the **player app** — with different purposes. Some things may be shared (e.g. API client, auth logic, types) where it makes sense.

---

## Stack

- **Backend** — Go. stdlib `net/http`, [slog](https://pkg.go.dev/log/slog), [envconfig](https://github.com/kelseyhightower/envconfig). See [Go style guide](./GO_STYLE.md).
- **Website** — **React + Vite + Bun**. One codebase for www.afterwave.fm (landing, discovery, account) and artist pages (handle.afterwave.fm). Built as a static (or pre-rendered) app; **served on a CDN**, preferably **AWS CloudFront** in front of S3. Client-side routing and API calls to api.afterwave.fm. **Different codebase** from the player app; different purpose (browse, discover, artist pages, account vs. play/download music).
- **Player app** — **Separate codebase** from the website. Purpose: play and download music, library, offline. **Web + desktop** share one React-based codebase within the player app; **Electron** wraps it for desktop (macOS, Windows, Linux). **Mobile** is **React Native + Expo** (iOS, Android) for background playing, lock screen, native audio. Same API and token model as the website; **different auth clients** per platform. Some code may be shared with the website (e.g. API client, types) where it makes sense. See [Player app](./PLAYER_APP.md).
- **Auth** — **Short-lived session tokens** and **long-lived refresh tokens**; linked in the DB. **Rolling refresh:** when a refresh token is used, it is **revoked** and the API returns a **new session token and a new refresh token** (one-time use → new pair). Tokens are **deleted when expired or when the refresh is used**. Different auth **clients** per platform: web (browser), desktop (Electron), iOS (React Native + Expo), Android (React Native + Expo). See [Sign-up and auth](./SIGNUP_AND_AUTH.md).

---

## Technical providers and services

| Provider / service | Use | Notes |
|--------------------|-----|--------|
| **Stripe** | Artist subscriptions (Billing), tips and fan→artist subs (Connect), platform fan sub. | See [Payments](./PAYMENTS.md), [Tax and compliance](./TAX_AND_COMPLIANCE.md). |
| **AWS** | Primary cloud: compute, API, DynamoDB, S3, CloudFront, Route 53, SES, Secrets Manager, KMS. | Design for **multiple global sites** (compute in several regions) with **one central database** (DynamoDB in one primary region, or DynamoDB Global Tables if we need multi-region read/write). See “AWS global” below. |
| **DynamoDB** | Primary database: users, artists, content metadata, sessions, refresh tokens, etc. | **Single table.** PK starts with **domain** (entity type). **KMS** for per-customer or sensitive-field encryption where required. See “DynamoDB: single-table design” below. |
| **S3** | Object storage: music files (full tracks, preview clips), images (covers, gallery, avatars). | **Definite.** Downloads and streaming URLs are **presigned CloudFront (or S3) URLs** issued by the API **only for signed-in users**; we do not stream bytes through the Go API. API validates auth → returns short-lived presigned URL → client hits CloudFront/S3 directly. |
| **CloudFront** | CDN for frontend static assets (React build). Also **signed URLs for media** (tracks, images) so downloads/streams go directly from the edge; API only issues URLs after auth. | Frontend: CloudFront → S3 (or origin). Media: CloudFront (or S3) with signed URLs; auth enforced at issue time by API. |
| **AWS SES** | Transactional email: signup, password reset, artist invites, notifications. | **Definite.** Sending domain and templates TBD. |
| **AWS Secrets Manager** | API secrets: JWT signing keys, Stripe keys, DB config, etc. Not in repo. | **Definite.** Fetched at startup or via IAM at runtime. |
| **KMS** | Encryption for **per-customer or sensitive data in DynamoDB** (e.g. payout details, tokens). | Use envelope encryption or DynamoDB encryption with customer-managed keys where we need to isolate or protect data per tenant/customer. |
| **OpenSearch (AWS)** | Search backend for discovery and feed: artist/user lookups and feed references live in **separate indices** (see below). Primary content stays in DynamoDB/S3; OpenSearch holds only what’s needed for search and feed ordering. | See [Discovery](./DISCOVERY.md), [Deployment](./DEPLOYMENT.md). |
| **Grafana Cloud** | Metrics, logs, and traces. We export OpenTelemetry (and optionally logs) from the API to Grafana Cloud to start. Can bring observability in-house later if cost becomes an issue. | See [Deployment](./DEPLOYMENT.md) for health, alerts, rollback. |
| **DNS** | Domain and wildcard: afterwave.fm, www.afterwave.fm, *.afterwave.fm, api.afterwave.fm. | Route 53 or equivalent. |

---

## OpenSearch: separate indices for discovery vs feed

We use **different OpenSearch indices (collections)** for discovery/artist-user lookups and for feed content. The actual content lives in DynamoDB (and S3 for blobs); OpenSearch holds only references and fields needed for search, filter, and sort.

| Index (collection) | Purpose | What’s stored |
|--------------------|---------|----------------|
| **Discovery** (e.g. `afterwave-discovery`) | Artist and user lookups, browse, search. | References and searchable/filterable fields: artist handle, display name, bio, genre, location, etc. No full content; primary data in DynamoDB. |
| **Feed** (e.g. `afterwave-feed`) | Feed listing and search over posts. | Post references and sort/search fields: `post_id`, `artist_handle`, `created_at`, optional body/excerpt for full-text search. Full post content and media pointers stay in DynamoDB/S3. |

- **Discovery** is used for: “search artists,” “artists in Bristol,” “browse by genre,” and any user-profile search we add. Queries return refs; the API loads full artist/user from DynamoDB when needed.
- **Feed** is used for: per-artist feed list, collated feed (“posts from artists I follow”), and search over post text. Queries return refs (e.g. `post_id` + `artist_handle`); the API loads full post (and media) from DynamoDB/S3 when needed.

Keeping these in separate collections avoids mixing feed documents with artist/user documents and lets us tune mappings, retention, and access per use case.

---

## AWS global: one central database, multi-region compute

- **Goal** — Support users and artists in multiple regions with low latency, while keeping **one central database** of users, artists, and content.
- **Database** — **Single DynamoDB** in one primary region (e.g. eu-west-1 for UK). All writes and authoritative reads go there. Option: **DynamoDB Global Tables** later if we need low-latency reads in other regions (replicated tables); for MVP, one region is enough.
- **Compute** — API can be deployed in **multiple regions** (e.g. eu-west-1, us-east-1) behind a global router (Route 53 latency-based or geo routing). Each region runs the same Go API; all regions talk to the **same DynamoDB** (in the primary region). So: global API endpoints, one DB. Cross-region DynamoDB latency is acceptable for API calls; we can add read replicas (Global Tables) later if needed.
- **S3 / CloudFront** — S3 bucket(s) in one region (or multi-region if we want); **CloudFront** in front for global distribution of frontend and media. Signed URLs for media work globally (CloudFront edge). No need to duplicate S3 per region for MVP.

---

## DynamoDB: single-table design

- **One table** for all entities: users, artists, sessions, refresh tokens, content metadata, etc.
- **PK starts with domain** (entity type). Pattern: `DOMAIN#...` where `DOMAIN` is the entity prefix (e.g. `USERS`, `ARTISTS`, `SESSIONS`, `REFRESH_TOKENS`).
- **Example (users):** `PK = USERS#user,{shortuuid[0]}` — domain `USERS`, then a partition component (e.g. first character of short UUID for distribution), then the full user ID or identifier as needed. SK (sort key) can be used for range queries or sub-entities (e.g. `USER#<id>` with `SK = PROFILE`, or `SK = EMAIL#<email>` for lookups). Exact PK/SK patterns per entity to be defined as we add features (artists, tracks, posts, etc.); the rule is **PK always starts with domain**.
- **Other domains (examples):** `ARTISTS#<handle>`, `SESSIONS#<session_id>`, `REFRESH_TOKENS#<token_id>`, etc. GSIs for alternate access patterns (e.g. user by email, artist by handle) as needed.
- **Short UUID** — Use a short UUID (or similar) for IDs; `shortuuid[0]` in the example is the first character of that ID, used in the partition key to spread load. Full ID in SK or in the same item as needed for lookups.

---

## Auth and tokens

- **Session token** — Short-lived (e.g. 15–60 minutes). Used for API calls (`Authorization: Bearer <session_token>`). Stored in memory or secure storage per client; not in cookies for API (optional cookie for web frontend only).
- **Refresh token** — Long-lived (e.g. days or weeks). Stored securely per client. Used only to obtain a new session token and a **new refresh token** (rolling refresh). **One-time use:** when a refresh token is used, it is **revoked**; the API returns a **new session token and a new refresh token**. Same user can have multiple refresh tokens (e.g. one per device); each use revokes that refresh and issues a new pair.
- **Linked tokens** — Session and refresh are linked in the DB. When refresh is used or either token expires, both are cleaned up (deleted or expired).
- **Auth clients** — We implement **separate auth clients** for **web**, **desktop**, **iOS**, and **Android**. Same API contract (login, refresh, logout); different storage and UX (e.g. OAuth redirect on web, secure enclave on iOS). Ensures tokens are handled correctly per platform (no shared secret between app types).

---

## Subdomains and artist handles (frontend only)

- **API does not use Host for routing.** All requests to the API go to api.afterwave.fm/v1/...; the API never looks at whether the request came from www or handle.afterwave.fm.
- **Frontend** — Served from the same CDN/origin for www and *.afterwave.fm (or same S3/CloudFront with routing rules). The frontend reads `window.location.hostname` (or equivalent): if it’s `www.afterwave.fm`, show landing or discovery; if it’s `something.afterwave.fm`, treat `something` as the artist handle, fetch `GET /v1/artists/something` (or similar), and render that artist’s page. Handle is **lowercase, no special characters**; display name is separate (stylised name for UI).

---

## Downloads and media (signed-in only, no stream through API)

- **Only signed-in users** can download full tracks or access media URLs. See [Sign-up and auth](./SIGNUP_AND_AUTH.md).
- **Flow** — Client requests a download or stream URL from the API (e.g. `GET /v1/tracks/:id/download-url`). API validates the user (session token), checks entitlement, then **returns a short-lived presigned CloudFront (or S3) URL**. Client uses that URL to **download or stream directly from CloudFront/S3**; the Go API never streams the bytes. This keeps bandwidth and cost off the API and uses CloudFront for global delivery.
- **Presigned URLs** — Short expiry (e.g. 5–15 minutes). Issued on demand. No need for “authenticated CloudFront” in the sense of validating every request at the edge; we validate once in the API and hand out a signed URL. Optional: CloudFront signed cookies or signed URLs with longer expiry for “stream this album” flows; TBD.

---

## Open decisions

- DynamoDB: exact PK/SK and GSI patterns per entity (users, artists, sessions, refresh tokens, content) as we add features; PK always starts with domain.
- Session and refresh token lifetime (minutes for session, days/weeks for refresh).
