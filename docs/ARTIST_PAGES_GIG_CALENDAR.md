# Artist pages: Gig calendar

*Design doc for the artist page gig calendar and the collated “all followed artists” calendar for signed-in users.*

See [Artist pages](./ARTIST_PAGES.md) for context. This is part of the core product that drives artist subscriptions.

## Implementation checklist

- Gig (event) fields: venue, location, date, time, timezone, link, notes
- Add/edit/delete: owner + "gigs" role; past gigs hidden or archived
- Display on artist page: Gigs section; chronological (soonest first)
- Collated calendar: signed-in users see all gigs from followed artists in one view
- API: list gigs per artist; list my (collated) gigs for feed/calendar page

---

## Purpose

- **On the artist page** — Show upcoming gigs/tour dates so fans can see when and where to see the artist live.
- **Collated calendar** — Signed-in users who follow artists can see a **single calendar view** of all gigs from artists they follow. This is a key feature: one place to see “what’s on” for everyone I follow.

---

## Gig (event) fields

Each gig is a single event. Suggested fields:

| Field | Purpose |
| ----- | ------- |
| **Venue** | Name of the venue (e.g. “The Fleece”). |
| **Location** | Place: city, country; optionally address or “city, state, country” for display. We can store structured (city, region, country) for discovery and timezone. |
| **Date** | Day of the gig (date only, or date + time — see below). |
| **Time** | Optional. Door time or show time; if omitted we show “TBA” or date only. |
| **Time zone** | Required if time is set (e.g. “Europe/London”). So the collated calendar can sort and display correctly for the user. |
| **Link** | Optional. Ticket URL, event page, or venue link. |
| **Notes** | Optional. Free text (e.g. “Support: TBA”, “All ages”). |

Recurrence (e.g. “every Tuesday”) is out of scope for v1; each gig is a one-off. We can add “duplicate to next week” or recurrence later.

---

## Add / edit / delete

- **Who** — Owner and invitees with “gigs” (or equivalent) permission can add, edit, and delete gigs.
- **Edit** — All fields are editable until the gig is in the past. We may hide or archive past gigs automatically (e.g. move to “Past gigs” or hide from the main list); TBD.
- **Delete** — Deleting a gig removes it from the page and from the collated calendar.

---

## Display on the artist page

- **Section** — The “Gigs” section (from the [site builder](./ARTIST_PAGES_SITE_BUILDER.md)) shows the list of upcoming gigs.
- **Order** — Chronological (soonest first). Past gigs can be hidden, shown in a separate “Past” list, or omitted; TBD.
- **Layout** — List or simple calendar strip. We don’t need a full month grid on the artist page for v1; a list with date + venue + link is enough. A calendar view can be an option or reserved for the collated view.

---

## Collated calendar (followed artists)

- **Who** — Signed-in users only. Requires “follow” relationship: the user follows the artist.
- **What** — A single view (e.g. “My gigs” or “Calendar”) that aggregates **all** upcoming gigs from **all artists the user follows**.
- **Content** — For each gig: date, time (with timezone or localised for the user), venue, location, **artist name** (so the user knows which artist), and link if present.
- **Order** — Chronological. User can filter by date range or artist if we support it later.
- **Where** — Lives on the main site (e.g. www.afterwave.fm/calendar or /me/calendar when logged in), not on each artist page.

This gives fans one place to see “what’s on” without visiting each artist page.

---

## Open decisions

- Whether to show past gigs on the artist page (and for how long).
- Exact timezone handling (store in UTC + IANA zone, display in user’s local time).
- Collated calendar: list vs calendar grid; filter by artist or date range.
- Recurring events in a later phase.
