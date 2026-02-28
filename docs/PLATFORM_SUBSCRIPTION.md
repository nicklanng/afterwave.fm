# Platform (fan) subscription

Optional subscription for users to support the platform. Distinct from subscribing to an artist. We use Stripe (same as [Payments](./PAYMENTS.md)); we receive this revenue.

## Implementation checklist

- Signed-in user can subscribe to Listener ($5) or Support ($10) tier (Stripe Billing)
- Free users get ads; Listener and Support tiers are ad-free and get 128 kbps Opus streaming
- Support tier: subscriber selects one artist to receive $5 of their subscription every month (we route $5 to artist, $5 to platform)
- Listener/Support badge (flair on profile, comments)
- Supporter blog: paid subscribers can publish blog on profile (reviews, links to artists)
- Personal photo gallery for paid subscribers (photos with links to artists)

---

## Tiers (locked)

| Tier | Price | Benefits |
|------|-------|----------|
| **Free** | $0 | Streaming with ads; 64 kbps Opus; monthly byte cap (e.g. 100 MB). |
| **Listener** | $5/month | No ads; 128 kbps Opus; no byte cap; badge, blog, personal gallery. |
| **Support** | $10/month | No ads; 128 kbps Opus; no byte cap; badge, blog, personal gallery; **subscriber selects one artist to receive $5 of their subscription every month** (we pay that $5 to the artist via Stripe Connect; platform keeps $5). |

---

## Who and what

- **Who** — Any signed-in user. Free users get ads; they can upgrade to Listener or Support to go ad-free and get higher quality.
- **Revenue** — Listener: we receive $5. Support: we receive $5, artist receives $5 (routed via Connect). Plus artist page subs. Free users fund the platform in part via ads (see [Indie-focused advertising plan](./INDIE_ADVERTISING_PLAN.md)).
- **Relationship to artist subs** — Separate product. A fan can be Listener (platform only), Support (platform + one artist gets $5/month), and/or subscribe to individual artists (recurring fan → artist is Patreon-style, via Stripe Connect; see [Payments](./PAYMENTS.md)).

---

## Listener / Support benefits

Subscribed users (Listener and Support) get **flair** and **features** that make the subscription feel worthwhile and that support the kind of community we want (zines, reviewers, photographers). Both tiers get the same streaming and ad-free benefits; Support adds the ability to direct $5/month to a chosen artist.

### Flair

- **Listener/Support badge** — Visible on their profile, comments, or wherever we show user identity. Signals “this person supports the platform.” We can add more flair options later (e.g. tier labels, custom badges).

### Blog

- **User blog** on their account — Paid subscribers can publish a blog on their profile: reviews, links to artists they’re listening to, recommendations, essays. Good for **zines**, **independent reviewers**, and people who want to write about music on the platform. Posts can link to artist pages so discovery flows both ways (artist → fan and fan → artist).

### Personal gallery

- **Personal photo gallery** — Paid subscribers can have their own gallery of photos they want to share, with **links to artists** on the platform. This ties into the [artist photo gallery](./ARTIST_PAGES_PHOTO_GALLERY.md) flow: users submit photos to artists (artist approves for the artist’s gallery); the same user can also showcase photos in their **own** gallery and link to the artist. Great for **photographers** and people who document shows — they get a presence on the platform and drive traffic to artists.

### Support tier: artist selection

- **Support tier only** — Subscriber chooses **one artist** to receive $5 of their $10 subscription every month. We route that $5 to the artist's Connect account (same as tips and recurring fan→artist subs). The subscriber can change their chosen artist at any time; the change applies from the next billing cycle (or we prorate; TBD). Artists see this in their dashboard as "Support-tier supporters" (who selected them). See [Artist pages: Support](./ARTIST_PAGES_SUPPORT.md).

---

## Summary

| Benefit | Purpose |
|--------|---------|
| Ad-free + 128 kbps Opus | Core listening upgrade for Listener and Support |
| Support: $5 to chosen artist | Direct platform revenue share with one artist per subscriber |
| Listener/Support badge | Flair; more options later |
| Blog | Reviews, recommendations, links to artists; zines, reviewers |
| Personal gallery | Photos with links to artists; ties to artist approval flow; photographers |

---

## Open decisions

- Billing interval (monthly locked; annual discount TBD).
- Whether blog and gallery are paid-only or available to all users (with paid getting extra capacity/visibility).
