# Music player app

Download-focused player for web, mobile, and desktop. We don’t stream through our API; we avoid ongoing streaming costs. Tracks are downloaded (or played via short-lived presigned URLs) for signed-in users only.

See [Architecture](./ARCHITECTURE.md) for API, auth, and presigned URLs. See [Sign-up and auth](./SIGNUP_AND_AUTH.md) for token model.

---

## Player app vs website (different codebases)

The **player app** and the **website** (www.afterwave.fm, artist pages) are **different codebases** with **different purposes**.

- **Website** — Landing, discovery, artist pages, account, support (tips, subs). React + Vite + Bun; served on CDN. One codebase for www and artist subdomains.
- **Player app** — Playing and downloading music, library, offline, presigned URLs. Separate codebase; purpose is listening and managing a music library, not browsing artist pages or account management.

**Some things may be common** (e.g. API client, auth logic, types) — shared via a shared package, or duplicated where it’s simpler. We share where it makes sense; we don’t force one codebase for both.

---

## Platforms and shared code (within the player app)

- **Targets** — **Web** (browser), **mobile** (iOS, Android), **desktop** (macOS, Windows, Linux). Within the player app we share code where it makes sense; we choose the right stack per platform for device support.
- **Web + desktop** — **One React-based codebase** for the player app on web and desktop. **Electron** wraps that React app for desktop (macOS, Windows, Linux). Same UI and business logic in the browser and in Electron. Tauri can replace Electron later if we want a lighter desktop bundle (optimization, not required for launch).
- **Mobile** — **React Native + Expo** for iOS and Android. Separate codebase from the player’s web/desktop React app, but same API and token model. We use React Native + Expo for **better device support**: background playing, lock screen integration, native audio controls. A web-view wrapper would share more code with web but would give weaker support for background audio and lock screen; we prioritise native behaviour for the player on mobile.
- **Reference** — Spotify uses a single React/TypeScript codebase for web and desktop; we do the same within the player app (React + Electron). For mobile, we use React Native + Expo for native audio and lock screen support.

---

## Auth clients (per platform)

We implement **different auth clients** for **web**, **desktop**, **iOS**, and **Android**. Same API contract (login, refresh, logout); different storage and UX:

- **Web** — Session and refresh in memory + secure storage (e.g. httpOnly cookie for refresh, or localStorage with care). Refresh flow in JS; redirect or fetch to api.afterwave.fm/v1/auth/refresh.
- **Desktop** — **Electron**: secure storage (OS keychain). Refresh token stored locally; session in memory. Native window for login if needed (OAuth or email/password).
- **iOS** — **React Native + Expo**: secure storage (Keychain via Expo SecureStore or native module). Refresh token in Keychain; session in memory.
- **Android** — **React Native + Expo**: secure storage (EncryptedSharedPreferences or Keystore via Expo SecureStore or native module). Same pattern as iOS.

Same token semantics everywhere: short-lived session, long-lived refresh; **rolling refresh** (use refresh → get new session + new refresh); linked tokens deleted when expired or when refresh is used. See [Sign-up and auth](./SIGNUP_AND_AUTH.md).

---

## Downloads and media (signed-in only)

- **Only signed-in users** can download full tracks or get stream URLs. See [Sign-up and auth](./SIGNUP_AND_AUTH.md).
- **Flow** — App calls API (e.g. `GET /v1/tracks/:id/download-url` or `GET /v1/tracks/:id/stream-url`) with session token. API validates user, checks entitlement, returns a **short-lived presigned CloudFront (or S3) URL**. App uses that URL to **download or stream directly from CloudFront/S3**; bytes do not go through the Go API. See [Architecture](./ARCHITECTURE.md).
- **Offline** — On mobile and desktop, app can **download** tracks to local storage and play offline. Download = fetch from presigned URL once, store on device. Library and playlists are synced via API (metadata); actual files are local after download.
- **Web** — In-browser: stream via presigned URL or temporary blob; optional “download for offline” using browser storage or PWA cache. Same presigned-URL contract.

---

## Relationship to web and API

- **Discovery** — User can discover artists on the **web site** ([www.afterwave.fm](http://www.afterwave.fm) or handle.afterwave.fm) or, if we add it, inside the player app (browse/search via API). They add artists/albums/tracks; the player app fetches metadata from API and download URLs when the user plays or downloads.
- **Auth** — Same user account as the website. Login in the player (web, desktop, iOS, Android) uses the same API (api.afterwave.fm/v1/auth/login, refresh, etc.) and the same token model. Different clients per platform; same identity.

---

## Open decisions

- How downloads are triggered (per track, per album, “download all from artist”).
- Format (e.g. MP3, lossless); quality tiers if any.
- Sync across devices (same account, same library metadata; files are local per device).
- Whether the player app and the website live in the same repo (monorepo with shared packages) or separate repos; and how much to share (API client, types, auth) vs keep separate.
- Tauri as a later swap for Electron if we want a lighter desktop bundle.

