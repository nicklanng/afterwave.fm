# Phasing and MVP

What we build first, what we defer. **Fill this in last** — after the rest of the product is designed — so we can decide MVP from a complete picture.

## Implementation checklist

- ~~Auth (sign-up, login, tokens, refresh, logout)~~
- ~~Artist page basics (create, get, update, delete; posts CRUD)~~
- ~~Following and collated feed (GET /feed)~~
- Payments (artist sub, tips, recurring, payouts)
- Discovery (location, genre, search)
- Player app; platform fan sub; advanced site builder; collated gig calendar; notifications
- MVP feature list and launch definition — open

---

## Scope

- **MVP / v1** — Minimum set of features to launch: which artist-page components (e.g. feed, music, gigs, photos?), discovery (location + genre?), auth, payments (artist sub, tips?). No need for everything at once.
- **Later phases** — e.g. player app, platform (fan) sub, advanced site builder, collated gig calendar, user-submitted photos.
- **Dependencies** — e.g. payments before tips; discovery before “artists near me” is useful.

---

## Open decisions

- Exact v1 feature list.
- Order of implementation (suggest: auth → artist page basics → payments → discovery → rest).
- Definition of “launch” (private beta, invite-only, public).
- Soft-launch with coupons for free artist accounts (until we implement subscriptions, so we can invite early artists to the platform to build up momentum and gather early feedback).