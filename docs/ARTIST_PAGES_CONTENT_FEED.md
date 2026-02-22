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

---

## Post types and content

- **Text** — The post body supports **Markdown** (bold, links, lists, headings, etc.). Clients should render the body as markdown when displaying posts. The API stores the body as a single string; no server-side markdown parsing or HTML generation.
- **Images** — We host images. Format/size TBD; one or multiple images per post.
- **YouTube embeds** — We do not host video. Artists can embed YouTube (or similar) by pasting a URL; we render the embed. Other embed providers (Vimeo, etc.) TBD.

We do not paywall posts; all posts are visible to everyone who views the artist page (subject to age-gating for explicit posts).

---

## Display and ordering

- **Chronological** — Feed is ordered by publish date, newest first (or oldest first if we support a toggle; default newest first).
- **On the page** — The “Feed” section (from the [site builder](./ARTIST_PAGES_SITE_BUILDER.md)) shows the list of posts. Layout: list or card layout; TBD.
- **Pagination or infinite scroll** — We support browsing older posts; exact UX TBD.

---

## Edit and delete

- **Who** — Owner and invitees with “feed” (or equivalent) permission can edit and delete posts they created; owner can edit/delete any post. TBD: whether “feed” role can edit/delete only their own or any post.
- **Edit** — Edits update the post in place; we may show “Edited” or last-edited time if we want transparency.
- **Delete** — Deleting a post removes it from the feed and from any notifications/history; we don’t expose deleted content.

---

## Open decisions

- Drafts (save without publishing) and scheduling (publish at a future time) — scope TBD; see [Artist pages → Content feed](./ARTIST_PAGES.md#content-feed).
- Multiple images per post: layout (gallery, carousel) and max count.
- Whether “feed” role can edit/delete only their own posts or any post on the page.
