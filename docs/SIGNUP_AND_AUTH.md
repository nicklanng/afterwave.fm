# Sign-up, login, and access control

Design for accounts, artist pages, and who can do what.

---

## Account model

- **We have users, not “artists” as an account type.** Everyone signs up as a user. An artist page is a resource a user creates.
- **Creating an artist page triggers the subscription.** Users can have an account for free; the paid subscription (~$10/month) starts when they create (and maintain) an artist page. One user can create or co-admin multiple artist pages; subscription model per artist page TBD (e.g. one sub per page vs one sub for multiple pages).

---

## Sign-up and login

- **Sign-up and login** are the same flow for all users (no separate “artist sign-up”).
- **Required for:** listening to full tracks, downloading, interacting (follow, comment, subscribe to artist, etc.), and creating or administering artist pages. One-off tips are the exception and do not require sign-in (see Tipping below).
- Mechanics (email/password, OAuth, magic link, etc.) to be decided in implementation; this doc stays at product level.

### Tokens (short-lived session, long-lived refresh)

- **Session token** — Short-lived (e.g. 15–60 minutes). Used for API calls (`Authorization: Bearer <session_token>`). Stored per client (memory or secure storage).
- **Refresh token** — Long-lived (e.g. days or weeks). Stored securely per client. Used only to obtain a new session token and a **new refresh token** (**rolling refresh**). **One-time use:** when a refresh token is used, it is **revoked**; the API returns a **new session token and a new refresh token**. Linked tokens are **deleted when expired or when the refresh is used**.
- **Linked in DB** — Session and refresh are linked. When either expires or refresh is consumed, both are cleaned up. Same user can have multiple refresh tokens (e.g. one per device); each use revokes that refresh and issues a new pair.
- **Auth clients** — We implement **different auth clients** for **web**, **desktop**, **iOS**, and **Android**. Same API contract (login, refresh, logout); different storage and UX per platform (e.g. secure enclave on iOS, secure storage on Android, browser storage or httpOnly cookie on web). See [Architecture](./ARCHITECTURE.md).

---

## Access control

### Viewing artist pages

- **Public.** Viewing an artist’s page (bio, feed, merch, track list, etc.) is **not** sign-up walled. Anyone can browse.

### Listening to music

- **Full listening and downloads require a signed-in user.** We do not allow anonymous users to stream or download full tracks (to avoid mass scraping).
- **Preview clips for anonymous visitors (optional):** we may offer short previews (e.g. ~10 seconds) on the artist page so anonymous visitors can sample without signing up. Full tracks remain behind login. Preview generation and hosting are an artist-page / media feature; behaviour is defined here so auth and media docs stay aligned.

### Tipping

- **One-off tips do not require sign-in.** Anonymous visitors can leave a one-off tip; to the artist it appears as **anonymous**.
- **Signed-in tipping is attributed.** If the user is signed in when they tip, the artist can see who it came from (e.g. username or display name).
- Recurring subscriptions to an artist, and subscribing to support the platform, still require a signed-in user.

### Other interacting and supporting

- **Interacting** with an artist (follow, comment, subscribe to artist, etc.) and **subscribing to support the platform** require a signed-in user.

---

## Artist page administration

- **Owner:** The user who creates the artist page is the owner. They pay the subscription and have full control (settings, billing, delete page, etc.).
- **Invited members:** The owner (and possibly other high-level admins) can **invite other users** by email or by username to help manage the artist page.
- **Configurable roles/permissions:** Invitees are assigned a role with a fixed set of permissions. Roles are configurable (we define the permission set; the owner picks which role to assign). Examples:
  - **Full admin / band member:** Everything the owner can do except maybe billing and “delete page” (TBD).
  - **Content: photos:** Can upload and manage photos only.
  - **Content: music:** Can upload and manage music only.
  - **Content: feed:** Can create and edit feed posts only.
  - **Custom combinations:** We can ship a set of predefined roles first and add “custom role” (pick permissions) later if needed.
- **Invitation flow:** Invitee gets an email/link; they must already be a user or sign up. Accepting grants access according to their role. Owner can revoke or change roles at any time.

---

## Summary


| Action                          | Requirement                                 |
| ------------------------------- | ------------------------------------------- |
| View artist page                | None (public)                               |
| Hear short preview (if we add it) | None                                      |
| Listen to full track / download | Signed-in user                              |
| One-off tip                     | None (shown as anonymous to artist)         |
| Tip while signed in             | Signed-in user (artist sees who gave)      |
| Follow, comment, subscribe to artist | Signed-in user                         |
| Subscribe to support platform   | Signed-in user                              |
| Create artist page              | Signed-in user + subscription               |
| Admin artist page               | Owner or invited user with appropriate role |


---

## Open decisions

- One subscription per artist page vs one subscription covering multiple pages per user.
- Exact list of permissions and predefined roles for invitees.
- Whether “full admin” invitees can invite/remove others or only the owner can.
- Preview clip length, format, and whether it’s opt-in per track/album or global per artist.

