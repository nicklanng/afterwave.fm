# Artist pages: Site builder

*Design doc for the artist page site builder. Start with basics; evolve toward a more capable (e.g. Squarespace-like) builder.*

See [Artist pages](./ARTIST_PAGES.md) for context. This is part of the core product that drives artist subscriptions.

## Implementation checklist

- Branding: display name, handle (lowercase), logo, cover image, accent colour(s)
- Sections: which sections show, in what order (Bio, Feed, Music, Photos, Gigs, Support, Links)
- Visibility: show/hide per section; default order for new pages
- Single-column stack in v1; no raw layout/columns yet
- Domain: subdomain {handle}.afterwave.fm; custom domain TBD
- Owner and invitees with "site" or "full admin" can edit builder

---

## Approach

- **Start minimal** — A small set of building blocks and options: which sections appear, in what order, and basic branding. No drag-and-drop or custom layouts at first.
- **Evolve later** — Over time we can add more layout control (e.g. column layouts, custom blocks, themes) toward a Squarespace-like experience. This doc defines the baseline.

---

## Branding (identity)

Artists own their identity. Configurable in the site builder / settings:

- **Display name** — Stylised name shown on the page (e.g. “Bare Naked Apology”). Can differ from the handle.
- **Handle** — Lowercase, no special characters; used in the URL `{handle}.afterwave.fm`. Set at creation; changing it may be allowed later with redirects (TBD).
- **Logo** — Image (avatar/logo) for the page. Format/size TBD.
- **Cover image** — Banner or hero image at the top of the page. Format/size TBD.
- **Colours** — Optional: accent colour(s) for links, buttons, or theme. We can start with one accent; full theme (background, text) later.

All of the above are editable by the owner and (as we define roles) by invitees with “site” or “full admin” permission.

---

## Sections (building blocks)

The artist page is composed of **sections**. Each section corresponds to a core component. Artists choose **which sections to show** and **in what order**.

| Section   | Content / purpose | Doc |
| --------- | ----------------- | --- |
| **Bio**   | Short text about the artist. Plain text or simple formatting at first; rich text TBD. | — |
| **Feed**  | Patreon-style content feed (posts). | [Content feed](./ARTIST_PAGES_CONTENT_FEED.md) |
| **Music** | Catalog of tracks/albums. | [Music](./ARTIST_PAGES_MUSIC.md) |
| **Photos** | Photo gallery (collections). | [Photo gallery](./ARTIST_PAGES_PHOTO_GALLERY.md) |
| **Gigs**  | Gig calendar (upcoming dates). | [Gig calendar](./ARTIST_PAGES_GIG_CALENDAR.md) |
| **Support** | Tips and “subscribe to artist.” | [Support](./ARTIST_PAGES_SUPPORT.md) |
| **Links** | External links (e.g. Bandcamp, social, merch store). List of label + URL. | — |

- **Visibility** — Artist can show or hide each section (e.g. hide “Gigs” until they have dates). Hidden sections are not rendered on the public page.
- **Order** — Artist can reorder sections (e.g. Music first, then Feed, then Gigs). Stored as an ordered list; we render top to bottom.
- **Defaults** — New artist pages can ship with a sensible default order (e.g. Bio → Feed → Music → Gigs → Photos → Support → Links). Artist can change it.

We do not expose raw “layout” (columns, sidebars) in v1; the page is a single-column stack of sections. Multi-column or custom layout is a later evolution.

---

## Domain and URL

- **Subdomain** — Artist pages live at `{handle}.afterwave.fm` (e.g. `barenakedapology.afterwave.fm`). The main site (www.afterwave.fm) is separate: discovery, account, billing, etc.
- **Custom domain** — TBD later phase. For now we only serve subdomains.

---

## Who can edit

- **Owner** — Full control: branding, section order, visibility, and all content.
- **Invitees** — Permissions are per-role. “Site builder” or “full admin” (or equivalent) can change branding and section order; content-only roles (e.g. “feed only,” “music only”) cannot. See [Sign-up and auth → Artist page administration](./SIGNUP_AND_AUTH.md#artist-page-administration).

---

## Open decisions

- Whether the handle can be changed after creation (and redirect from old URL).
- Logo/cover dimensions, file types, and max size.
- Rich text for bio (or keep plain text in v1).
- When to introduce multi-column or advanced layout.
