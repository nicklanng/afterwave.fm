# Artist pages: Content feed

*Design doc for the artist page content feed (Patreon-style posts: text, images, YouTube embeds).*

See [Artist pages](./ARTIST_PAGES.md) for context. This is part of the core product that drives artist subscriptions.

---

## Post creation (upload flow)

- Owner and invitees with “feed” (or equivalent) permission can create posts (text, images, YouTube embeds; we don’t host video).
- Drafts, scheduling, rich text — scope TBD.

### Guidelines and explicit on the post form

The **new post** (content feed) form includes:

1. **Guidelines section** — A short summary of our [content policy](./MODERATION.md): human-made only; no illegal content, harassment, hate speech, impersonation, or porn; you must have the rights to what you post. Link to full guidelines. Shown on the same screen as the post form (e.g. above or beside the form, or in a collapsible “Guidelines” block).
2. **Explicit option** — A checkbox or toggle: **“Mark as explicit”** (mature themes: strong language, sex, drugs, violence, etc.). When set, the post is age-gated: only users who have confirmed they are 18+ can view it (or we hide/blur it until age-verified). Artists decide; we don’t censor. See [Moderation → Minors, explicit content](./MODERATION.md#minors-explicit-content-and-child-safety).

These elements are shared with other upload flows; see [Artist pages → Upload guidelines and explicit content](./ARTIST_PAGES.md#upload-guidelines-and-explicit-content).
