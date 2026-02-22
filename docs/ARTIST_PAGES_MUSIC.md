# Artist pages: Music

*Design doc for the artist music catalog: uploads, track limits per tier, full listen/download for signed-in users, preview clips for anonymous.*

See [Artist pages](./ARTIST_PAGES.md) for context. This is part of the core product that drives artist subscriptions.

---

## Upload flow (tracks / albums)

- Owner and invitees with “music” (or equivalent) permission can upload tracks and organise them into albums.
- **Track limits** — Per [Payments](./PAYMENTS.md): subscription tier determines max tracks (25 / 100 / unlimited). We block further uploads when the limit is reached and surface upgrade options. The limit counts **unique tracks**, not how many times a track appears in releases. If an artist uploads a few tracks as a single and later releases an album that includes those same tracks, the album can **reuse the same upload slots** — they don't have to upload the tracks twice.
- Format/specs (e.g. file types, max size, metadata) TBD.

### Guidelines and explicit on the upload page

The **music upload** page includes:

1. **Guidelines section** — A short summary of our [content policy](./MODERATION.md): human-made only; no illegal content, harassment, hate speech, impersonation, or porn; you must have the rights to the music. Link to full guidelines. Shown on the same screen as the upload form (e.g. above or beside the form, or in a collapsible “Guidelines” block).
2. **Explicit option** — A checkbox or toggle: **“Mark as explicit”** (mature themes: strong language, sex, drugs, violence, etc.). When set, the track (or album) is age-gated: only users who have confirmed they are 18+ can stream or download it. Artists decide; we don’t censor. See [Moderation → Minors, explicit content](./MODERATION.md#minors-explicit-content-and-child-safety).

These elements are shared with other upload flows; see [Artist pages → Upload guidelines and explicit content](./ARTIST_PAGES.md#upload-guidelines-and-explicit-content).

---

## Format and specs (upload)

- **File types** — We accept common lossless or high-quality formats (e.g. WAV, FLAC, or high-bitrate MP3). Exact list and preferred format TBD; we may transcode for streaming and download.
- **Max file size** — Per track; TBD (e.g. 500 MB per file or similar).
- **Metadata** — Title, artist (defaults to page name), optional album, track number, year, cover art. We store and display this; exact fields TBD.
- **Albums** — Tracks can be grouped into albums (or “releases”). An album has a title, optional cover, and ordered list of tracks. The same track can appear in multiple releases (e.g. single + album) without counting twice toward the track limit; see [Payments → Track limit](./PAYMENTS.md).

---

## Display on the artist page

- **Section** — The “Music” section (from the [site builder](./ARTIST_PAGES_SITE_BUILDER.md)) shows the catalog.
- **Organization** — Artists can present releases (albums, EPs, singles) and/or a flat track list. We show releases with cover art and track list; play/download per track or per release TBD.
- **Order** — By release date (newest first) or manual order; TBD.

---

## Listening and download

- **Signed-in users** — Full streaming and download. No paywall; all tracks are free for signed-in users.
- **Anonymous visitors** — No full tracks (to avoid scraping). Optional: **preview clips** (e.g. ~10 seconds) so visitors can sample without signing in. Preview length, format, and whether it’s opt-in per track/album or global per artist are open; see [Artist pages → Open decisions](./ARTIST_PAGES.md#open-decisions) and [Sign-up and auth → Preview clips](./SIGNUP_AND_AUTH.md#listening-to-music).

---

## Integration with player app

The [music player app](./PLAYER_APP.md) is download-focused (phone and desktop). How the app discovers and downloads tracks from the artist catalog (e.g. via API, same account) is to be defined when we implement the player; this doc assumes the catalog is the source of truth for tracks and metadata.

---

## Open decisions

- Exact file types, max size, and transcoding strategy.
- Preview clips: length (e.g. 10 s), format, opt-in scope (per track vs global).
- Whether we support “reuse same track in multiple releases” from day one (counts once toward limit) or add later.
