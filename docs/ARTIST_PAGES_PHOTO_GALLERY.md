# Artist pages: Photo gallery

*Design doc for the artist page photo gallery.*

See [Artist pages](./ARTIST_PAGES.md) for context. This is part of the core product that drives artist subscriptions.

## Implementation checklist

- Collections: create, name, optionally link to gig event; optional "post to feed" on create
- Artist uploads: owner + "photos" role; bulk upload; ordering within collection
- Guidelines (and optional mature/explicit) on upload form
- User-submitted photos: link to event; artist approval before public
- Display: grid or masonry; lightbox/full-size; tie-in with supporter personal gallery TBD

---

## Collections

- Artists upload photos organised into **collections** (e.g. “Tour 2024”, “Studio sessions”, “Press”).
- Artists can create a collection **linked to an event from their gig calendar** — e.g. “Photos from Bristol, 12 March” tied to that gig. This keeps event photos grouped and discoverable.
- **Post to feed** — When creating a new collection, the artist has an option to **post it to their content feed**. If chosen, a post is created (e.g. “New gallery: [collection name]” with a link or preview to the collection) so followers see the new collection in their feed. Optional at creation time; artists can also share collections manually via the feed later.

---

## Artist uploads

- Owner and invitees with “photos” (or equivalent) permission can create collections and upload photos.
- **Bulk upload** — Artists can upload multiple photos at once into a collection; max count per batch and per collection TBD.
- **Ordering** — Photos within a collection can be reordered (drag-and-drop or manual order). Default order: upload order or by date; TBD.
- **Display** — Layout (grid, masonry, carousel) and display behaviour (lightbox, full-size view) TBD. We start with a simple grid; evolve as needed.

### Guidelines (and optional explicit/mature) on the upload form

The **photo upload** flow includes a **guidelines section**: short summary of our [content policy](./MODERATION.md) (human-made only; no illegal content, harassment, hate speech, impersonation, or porn; you must have the rights). Link to full guidelines. Optionally we add a **“Mark as mature”** or **“Explicit”** control for artistic nudity or sensitive imagery so it can be age-gated; TBD. See [Artist pages → Upload guidelines and explicit content](./ARTIST_PAGES.md#upload-guidelines-and-explicit-content).

---

## User-submitted photos

- **Users can submit photos to artists** (e.g. fan shots from a gig). They can link a submission to an **event** on the artist’s calendar.
- **Artists must approve** user-submitted photos before they appear in the gallery. Until approved, submissions are in a pending/queue state visible only to the artist (and admins with the right role).
- Approved user photos can be placed in the same event-linked collection or a dedicated “Fan photos” (or similar) collection — TBD.

**Tie-in with supporter galleries:** Users who have a [platform subscription](./PLATFORM_SUBSCRIPTION.md) can also have a **personal gallery** on their profile, with photos linked to artists. The same photo might be submitted to an artist (for the artist's gallery, pending approval) and shown in the user's own gallery with a link to the artist. Good for photographers and reviewers.

---

## Open decisions

- Whether approved user photos live in the same collection as the event or in a separate “fan photos” area.
- Permissions: who on the artist side can approve/reject submissions (owner only vs any “photos” role).
- Limits on submissions per user per event, file size, etc.
