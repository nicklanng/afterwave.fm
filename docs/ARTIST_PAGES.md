# Artist pages

What an artist page is, what’s on it, and how it works. For who can view or manage it, see [Sign-up and auth](./SIGNUP_AND_AUTH.md).

**Component design docs:** Each core component below has its own design document. This is the main product that drives artist subscriptions; deeper design lives in those docs.

## Implementation checklist

- ~~Artist page model: user creates page; public view; handle.afterwave.fm~~
- ~~Artists API: create, get by handle, update, delete; posts (feed) CRUD~~
- ~~Collated feed (GET /feed) for signed-in users from followed artists~~
- Site builder: sections, order, branding (display name, handle, logo, cover, colours)
- Content feed: posts (text, images, YouTube); guidelines + explicit flag on form
- Notifications: opt-in "notify me"; bell + badge; collated feed; no email export to artists
- Music: uploads, track limits per tier; full listen/download signed-in; optional preview for anonymous
- Photo gallery: collections; event-linked; user submissions with artist approval
- Gig calendar: upcoming gigs on page; collated calendar for followed artists
- Support: tips (anonymous/attributed), recurring subscribe to artist; merch links only
- Upload guidelines and explicit option on music, feed, photo flows
- Invited members and configurable roles (owner, full admin, content roles)

---

## What an artist page is

An **artist page** is the main presence for an artist (band, solo act, etc.) on Afterwave.fm. A user creates it and pays the platform subscription; they own it and can invite other users to help run it with configurable roles. The page is **public** — anyone can view it without signing in. Full music listening and downloads require a signed-in user; one-off tips can be anonymous.

**URLs:** Artists get subdomains like `barenakedapology.afterwave.fm`. The **handle** (band ID) is **lowercase, no special characters**; artists can supply a **stylised name** for display (e.g. handle `barenakedapology`, display name “Bare Naked Apology”). Custom (own) domain TBD later.

The product promise (from [Vision](./VISION.md)): site builder, content feed, notifications, music, photos, gigs, and support — in one place. All content is free to access; artists are supported by tips, subscriptions, and gigs (and by selling merch themselves); we take no cut of that income.

**Blocking** — Signed-in users can [block artists](./DATA_AND_PRIVACY.md#blocking-artists-and-other-users); blocked artists’ pages (and all their content) are hidden from the blocker. We don’t show the page or link to it in discovery or elsewhere for that user.

---

## Core components

These are the main parts of the artist page. Each can have its own design document for detailed behaviour and UX.

| Component | Purpose |
| --------- | ------- |
| **Site builder** | Shape the page: layout, sections, branding. Start with basics; evolve toward a more capable (e.g. Squarespace-like) builder. [Design doc](./ARTIST_PAGES_SITE_BUILDER.md). |
| **Content feed** | Patreon-style feed of posts (text, images, YouTube embeds — we don’t host video). All visible for free. [Design doc](./ARTIST_PAGES_CONTENT_FEED.md). |
| **Notifications** | “Be notified of new content on this artist page.” Email is optional; core is opt-in to notifications. [Design doc](./ARTIST_PAGES_NOTIFICATIONS.md). |
| **Music** | Catalog of tracks/albums. Track limit per subscription tier: 25 / 100 / unlimited (see [Payments](./PAYMENTS.md) for tiers). Full listen/download for signed-in users; optional short previews for anonymous. [Design doc](./ARTIST_PAGES_MUSIC.md). |
| **Photo gallery** | Photos in collections; collections can link to gig calendar events. Users can submit photos (linked to events); artists approve before they appear. [Design doc](./ARTIST_PAGES_PHOTO_GALLERY.md). |
| **Gig calendar** | Upcoming gigs/tour dates. Signed-in users who follow artists see a **collated calendar** of all gigs from artists they follow. [Design doc](./ARTIST_PAGES_GIG_CALENDAR.md). |
| **Support / income** | Tips (anonymous or attributed), optional recurring subscriptions to the artist. We don’t take a cut. [Design doc](./ARTIST_PAGES_SUPPORT.md). |
| **Merch** | For now: artists sell and handle orders themselves (e.g. link out to their store). We do not act as supplier or distributor. If artist demand appears later, we can revisit. |

---

## Site builder

- Artists customise how their page looks and what sections appear (bio, feed, music, photos, gigs, links, etc.).
- **Approach:** Start with a minimal set of building blocks and options; slowly build toward a more evolved, Squarespace-style builder. The basics we can do first need their own design doc.
- Branding: name, handle/slug (used in `handle.afterwave.fm`), logo, cover image, colours — artist owns their identity.
- **Domains:** Artist pages live at `{handle}.afterwave.fm` (e.g. `barenakedapology.afterwave.fm`).

---

## Content feed

- Chronological feed of posts: text, images, **YouTube embeds** for video. We do not host video at this point.
- No paywalling — all posts visible to everyone who views the page.
- Who can post: owner and any invited member with “feed” (or equivalent) permission.
- Features such as drafts, scheduling, rich text — scope TBD.

---

## Notifications (“mailing list”)

- **Core idea:** “Be notified of new content on this artist page.” Fans opt in to get notified when the artist posts new content (and optionally when there’s new music, gigs, etc.).
- **Main page:** Signed-in users see a **collated feed** of feed updates from the artists they follow (one stream on the main site). Notifications are accessed via a **bell**; the bell shows a **badge** when there are unread notifications, and the user clicks the bell to open the notification list. See [Artist pages: Notifications](./ARTIST_PAGES_NOTIFICATIONS.md).
- **Email is optional.** We may offer email delivery as an option. We **don’t expose subscriber emails to artists** — we (the platform) own the sign-up data and send notifications on the artist’s behalf; artists see counts and can trigger “notify subscribers” through us. See [Data and privacy](./DATA_AND_PRIVACY.md).

---

## Music on the artist page

- Artists upload tracks/albums. All are free to listen and download for **signed-in** users.
- **Track limits by subscription tier:** Limits are per subscription level, e.g. 20 tracks, 60 tracks, 100 tracks, unlimited. Exact numbers and subscription pricing to be decided.
- **Anonymous visitors:** no full tracks (to prevent scraping). Optional: short preview clips (e.g. ~10 seconds) so visitors can sample without signing in.
- Format/specs for uploads and how the download-focused player app uses this catalog — to be defined in the music design doc.

---

## Photo gallery

- Artists upload photos in **collections**. They can create a collection linked to an **event from their gig calendar** (e.g. photos from a specific gig).
- **User-submitted photos:** Fans can submit photos to the artist and link them to an event. Artists must **approve** submissions before they appear in the gallery.
- Permissions: owner and invitees with “photos” (or equivalent) role. Layout, bulk upload, ordering — see [Photo gallery design doc](./ARTIST_PAGES_PHOTO_GALLERY.md).

---

## Gig calendar

- Artists add upcoming gigs/tour dates to their page. Shown on the artist page as a calendar/list.
- **Collated calendar for signed-in users:** Users who follow artists can see a single calendar view of **all** gigs from artists they follow. This is a key feature of the gig component.
- Fields, recurrence, time zones, and UX to be specified in a dedicated design doc.

---

## Support and income (on the page)

- **Tips:** One-off tips; no sign-in required. Anonymous tips show as “anonymous” to the artist; signed-in tips are attributed. We don’t take a cut; payment processor fees only.
- **Recurring subscriptions to the artist:** Optional; fan pays periodically to support the artist. Requires signed-in user. All revenue to artist (we only take payment-processor fees).
- **Merch:** Artists sell and handle orders themselves (links, own store, etc.). We do not act as supplier or distributor for now; we can revisit if there’s demand from the artist user base.

---

## Upload guidelines and explicit content

Every **artist upload** flow (music, feed posts, photos) includes a **guidelines** section and, where applicable, an **option to set content as explicit**.

- **Guidelines section** — On the upload page (or a collapsible “Guidelines” block), we show a short summary of what’s allowed and what isn’t: human-made content only; no illegal content, harassment, hate speech, or impersonation; no porn or sexual smut; you must have the rights to what you upload. We link to the full [content policy / moderation](./MODERATION.md) so artists can read the full rules. This appears on: **music upload**, **new post** (content feed), and **photo upload** (and any other future upload surfaces).
- **Explicit option** — For **music** (tracks/albums) and **feed posts**, we provide a control to **mark content as explicit** (mature themes: strong language, sex, drugs, violence, etc.). Explicit content is then age-gated so only users who have confirmed they are 18+ can stream or download it. Artists choose the flag; we don’t censor. For **photos**, we may offer a “mature” or “explicit” option for artistic nudity or sensitive imagery so it can be age-gated; TBD. See [Moderation → Minors, explicit content, and child safety](./MODERATION.md#minors-explicit-content-and-child-safety).

Implementation: the guidelines can be a short, always-visible blurb (e.g. “By uploading you agree to our [content guidelines]. No illegal content, harassment, hate speech, or porn.”) with a link, plus an **explicit** checkbox or toggle on the same form. We do not require artists to click “I have read the full policy” to submit; we rely on the visible summary and terms of service.

---

## Administration and roles

- **Owner** creates the page, pays the subscription, has full control.
- **Invited members** get roles with specific permissions (e.g. full admin, photos only, music only, feed only). See [Sign-up and auth → Artist page administration](./SIGNUP_AND_AUTH.md#artist-page-administration).

---

## Open decisions

- Track limits and subscription tiers are defined in [Payments](./PAYMENTS.md) (25 / 100 / unlimited).
- Preview clips: length, format, opt-in scope.
- Notifications: in-app only at first, or email from day one; export for artists.
- Custom (own) domain for artist pages — later phase.