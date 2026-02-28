# Afterwave.fm — Vision

## Implementation checklist

- ~~Core API and auth (sign-up, login, tokens, artists, feed, following)~~
- Artist pages: site builder, feed, notifications, music, photos, gigs, support (see [Artist pages](./ARTIST_PAGES.md))
- Discovery: location, genre, search, ranking
- Player app: download-focused; web, desktop, mobile
- Payments: artist subscription, tips, recurring fan→artist, platform fan sub, payouts
- No paywall on content; voluntary support only; no cut of artist income

---

## Product Overview

Afterwave.fm is a competitor to **Spotify** and **Bandcamp**, and operates in the broader space of music stores and artist platforms. The core difference: **nothing is paywalled**. All music, posts, and content are free for everyone. Artists are supported by voluntary donations, subscriptions, merch, and gigs — not by gatekeeping access.

**For artists:** A ~$10/month subscription buys a site builder, mailing list, and content feed (Patreon-style). They keep full ownership of their audience, branding, and income. The platform takes **no cut** of artist income — only the software subscription fee (plus standard payment-processor fees, e.g. Stripe).

**For listeners:** Sign-up is required only to interact with artists or download music; listening and downloading remain free. Free users may see light ads. Fans can optionally subscribe: **Listener** ($5/month, no ads, 128 kbps Opus) or **Support** ($10/month, no ads, 128 kbps Opus, plus select one artist to receive $5 of their subscription every month).

**Listening experience:** A dedicated music player app (phone and desktop) focuses on **downloading** music rather than streaming, avoiding ongoing streaming infrastructure costs.

---

## Platform Manifesto

### Core Philosophy

- **Human-made music only.** This platform exists to support real musicians creating intentional art.
- **Music is never paywalled.** All music is freely accessible to listen and download. No gated tracks, no locked albums, no forced minimum pricing.
- **Support is voluntary.** Financial contribution comes from tips, subscriptions, merch, and gigs — not access control.
- **Discovery over distribution.** We are not a streaming farm. We are a discovery engine for independent scenes.
- **Artist sovereignty first.** Artists own their audience, mailing list, branding, and income streams.
- **Scene density over global sprawl.** It is better to matter deeply within real communities than to exist vaguely everywhere.
- **Lean infrastructure, not scale-at-all-costs.** Sustainability over hype.
- **No artificial promotion systems.** Music rises through genuine discovery, not paid positioning.
- **Build tools musicians actually use.** Replace fragmented band tools with one clean system.

### Revenue Principles (Non-Negotiables)

- We do not take a percentage of artist income.
- All Pay-What-You-Want revenue goes directly to the artist.
- Platform sustainability comes from software subscription, not artistic cuts.
- Clear, flat pricing. No hidden fees.
- No surprise bandwidth billing for small artists.
- No monetisation via selling user data or targeted advertising; light ads for free users only to fund the platform (Listener/Support tiers are ad-free).
- No growth incentives that distort artistic fairness.
- No venture capital pressure that compromises product direction.

**Primary revenue model:**

- Flat subscription for bands.
- Optional Listener ($5) and Support ($10) subscription for fans; free users may see light ads.
- Premium tooling tiers — never revenue-sharing exploitation.

### Team Size Cap

- Stay under 7 full-time employees until well past significant profitability.
- Prefer high-output generalist engineers over layered departments.
- No unnecessary management layers.
- No bloated middle roles without clear output.
- All team members must directly contribute to product quality or community health.
- Fair pay, lean org chart, low politics.

*Small team, high ownership, no bureaucracy.*

### What We Refuse To Add

- No algorithmic feed designed to maximise engagement addiction.
- No AI-generated music hosting.
- No label promotion deals.
- No paid boost placements.
- No ad-based revenue from selling user data; light ads for free users only (paid tiers ad-free).
- No dark UX patterns pushing upgrades.
- No tokenised hype economy.
- No venture capital mandates that override culture.
- No forced DRM or invasive control systems.
- No growth experiments that degrade artistic trust.

### Success Definition

- Culturally relevant in real independent scenes.
- Financially sustainable without exploitation.
- Operable by a small, focused team.
- Trusted by musicians.
- Loved by listeners.
- Not dependent on endless funding rounds.

---

*Music is culture, not a commodity. The commodity is the software infrastructure.*
