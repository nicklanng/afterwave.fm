# Artist pages: Music

*Design doc for the artist music catalog: uploads, track limits per tier, full listen/download for signed-in users, preview clips for anonymous.*

See [Artist pages](./ARTIST_PAGES.md) for context. This is part of the core product that drives artist subscriptions.

---

## Upload flow (tracks / albums)

- Owner and invitees with “music” (or equivalent) permission can upload tracks and organise them into albums.
- **Track limits** — Per [Payments](./PAYMENTS.md): subscription tier determines max tracks (25 / 100 / unlimited). We block further uploads when the limit is reached and surface upgrade options.
- Format/specs (e.g. file types, max size, metadata) TBD.

### Guidelines and explicit on the upload page

The **music upload** page includes:

1. **Guidelines section** — A short summary of our [content policy](./MODERATION.md): human-made only; no illegal content, harassment, hate speech, impersonation, or porn; you must have the rights to the music. Link to full guidelines. Shown on the same screen as the upload form (e.g. above or beside the form, or in a collapsible “Guidelines” block).
2. **Explicit option** — A checkbox or toggle: **“Mark as explicit”** (mature themes: strong language, sex, drugs, violence, etc.). When set, the track (or album) is age-gated: only users who have confirmed they are 18+ can stream or download it. Artists decide; we don’t censor. See [Moderation → Minors, explicit content](./MODERATION.md#minors-explicit-content-and-child-safety).

These elements are shared with other upload flows; see [Artist pages → Upload guidelines and explicit content](./ARTIST_PAGES.md#upload-guidelines-and-explicit-content).
