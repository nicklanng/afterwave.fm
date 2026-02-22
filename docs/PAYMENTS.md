# Payments and payouts

How money moves: artist subscription fee, tips, recurring artist subs, platform (fan) subscription, and payouts to artists.

---

## Artist subscription (platform fee)

Artists pay a flat monthly fee to have an artist page. We receive this; it’s our primary revenue (per [Vision](./VISION.md)). Tiers are below.

### Proposed tiers (locked for design)


| Tier          | Price      | Track limit      | Notes                                                        |
| ------------- | ---------- | ---------------- | ------------------------------------------------------------ |
| **Base**      | $10/month  | 25 tracks        | Artists can remove and add tracks freely (rotate catalog).   |
| **Growth**    | $20/month  | 100 tracks       | For bigger catalogs.                                         |
| **Unlimited** | $50/month  | Unlimited tracks | For large back catalogs or artists who want no cap. |


- **Important:** All tiers get the same **core features** — feed, gallery, gigs, notifications, tips, recurring fan subs, mailing list, discovery. We do **not** lock $10 users out of anything essential. Tiering is by **capacity** (tracks, and maybe photo storage / custom domain later), not by “can you do X.”
- **Track limit = unique tracks.** The limit counts distinct tracks, not how many times a track appears in releases. If an artist uploads a few tracks as a single and later releases an album that includes those same tracks, the album reuses the same upload slots — they don't have to upload the tracks twice. See [Artist pages: Music](./ARTIST_PAGES_MUSIC.md).
- **Other perks** for higher tiers (optional, to add later): e.g. more photo storage, custom domain, deeper analytics, priority support, or multiple artist pages per account. Nothing that makes $10 feel second-class for day-to-day use.

### Competitor context (how we compare)

- **Bandcamp** — No monthly fee for artists; they take ~15% of digital sales (drops to 10% after a threshold) and ~10% on merch. So artists pay nothing to have a page, but they give up a cut of every sale. We’re the opposite: flat monthly, **zero** cut of artist income. For artists who sell a lot, Bandcamp’s % can exceed our $10–50/month; for artists who rely on tips/subs and don’t want to share that, we’re cheaper and clearer.
- **SoundCloud** — Pro tiers (~$12–16/month) for more upload time, distribution, stats. They’re a streaming/discovery platform first; we’re a “your own page + discovery” play. Our $10 is in the same ballpark for “presence + features.”
- **Patreon** — Creators don’t pay to have a page; Patreon takes 5–12% of what patrons pay. Again, we’re flat fee, no cut of artist income.
- **Squarespace (or similar site builders)** — ~$16–23/month for a website. We’re not just a site builder; we’re music-focused (tracks, discovery, tips, fan subs). $10 is competitive for “your artist presence + tools in one place.”

**Takeaway:** Our $10 base is competitive and fair. The “no cut of artist income” message is the differentiator; tiering by track count (and capacity) keeps the offer simple and doesn’t gate core product.

### Honest feedback on the tiers

- **$10 for 25 tracks** — Strong. Enough for 2–3 albums or a solid set of singles; “remove and add whatever you want” avoids lock-in and fits how artists rotate releases. Good entry point; aligns with competitor entry tiers (~$10–16).
- **$20 for 100 tracks** — Good middle. 2× the price for 4× the tracks; artists with a real back catalog get room. Sits between SoundCloud-style mid tiers and Squarespace.
- **$50 for unlimited** — 2.5× the price for no cap. Smaller jump than $100; easier to justify for artists who want no limits. Add perks later (custom domain, analytics) so the tier isn’t only track count. Don’t strip $10 of anything important — capacity and nice-to-haves only.

### Billing behaviour

- **Monthly**, charged on a **fixed date each month**.
- **Anniversary billing:** Charge on the same calendar day they subscribed (e.g. sign up on the 15th → charge on the 15th each month). If the month has fewer days (e.g. 31st → February), we use the last day of the month. We aim to be as consistent as possible so artists know when to expect the charge.
- Stripe supports this natively (subscription with `billing_cycle_anchor` or equivalent).

### Provider

We use **Stripe** for artist subscriptions (Stripe Billing) and for tips and payouts to artists (Stripe Connect). Tax and compliance as merchant of record are in [Tax and compliance](./TAX_AND_COMPLIANCE.md).

---

## Recurring subscriptions to artist (fan → artist)

Fans can subscribe to an artist (Patreon-style): they pay periodically (e.g. monthly), and **100% goes to the artist** — we don’t take a cut. We use **Stripe Connect**: the fan pays through us; Stripe routes the funds to the artist’s connected account. The payment processor fee is **deducted from the artist’s payout**; we are upfront with artists about handling fees (see [Payouts to artists](#payouts-to-artists)).

- **Who** — Signed-in user subscribes to an artist; artist must have a connected Stripe account to receive.
- **What we do** — Facilitate the subscription (Connect); we do not take a percentage. See [Platform subscription](./PLATFORM_SUBSCRIPTION.md) for platform (fan) sub benefits; artist subs are separate.

---

## Platform (fan) subscription

Optional subscription to support the platform. We receive this; Stripe Billing. See [Platform subscription](./PLATFORM_SUBSCRIPTION.md).

---

## Tips (one-off)

Fans can leave a one-off tip for an artist. Anonymous (no sign-in) or attributed (signed-in). 100% to the artist (we don’t take a cut); the payment processor fee is **deducted from the artist’s payout**. We are upfront with artists about handling fees. Same Connect flow as payouts below — tip goes to artist’s connected account.

---

## Payouts to artists

We pay artists the money they’re owed from **tips** and **recurring fan → artist subscriptions**. We use **Stripe Connect**: artists connect their Stripe account (or we create a Connect account for them); when a fan tips or pays a recurring sub, Stripe collects the payment and we transfer the balance to the artist’s connected account.

### What we pay out

- **Tips** — One-off payments from fans (anonymous or attributed). Full amount minus any processor fee (see fee handling below).
- **Recurring subscriptions** — Monthly (or other interval) payments from fans who subscribe to that artist. Full amount minus any processor fee.

We do **not** take a platform percentage. Per [Vision](./VISION.md), all of this revenue goes to the artist; we only receive our own revenue (artist page sub, platform fan sub).

### Connect account types (artists can choose)

Stripe Connect supports three account types. We support **Standard** and **Express** so artists can choose; **Custom** is not for our scale.

- **Standard** — Artist has (or creates) their **own Stripe account** and links it to us via OAuth. They have full control, their own Stripe dashboard, and a direct relationship with Stripe. Best for artists who already use Stripe or want full visibility. No extra Connect fee; standard processing fees only.
- **Express** — We create a **Stripe-hosted** connected account for the artist. Stripe runs onboarding and handles identity/KYC; we can customize the experience. Artist gets a simplified Stripe experience (e.g. Express Dashboard). Good for artists who want the fastest setup. Stripe may charge a small per-account fee.
- **Custom** — Platform builds everything (white-label). Highest engineering and compliance effort; not for us at current scale.

**Artists choose:** During setup we offer “Link your existing Stripe account” (Standard) or “Set up with Stripe through us” (Express). We document both in help/terms.

**When they connect:** Setting up or linking Connect is part of the **artist (page) registration flow** — we prompt them to connect so they can receive tips and recurring subs. They can **bypass** Connect if they **disable financial support** from fans (no tips, no recurring subs). In that case they still pay the platform subscription for the artist page; they just don’t accept fan payments.

### How Connect works (in principle)

1. **Artist onboarding** — Artist completes Connect setup (Standard or Express) as part of artist registration, or skips it by disabling financial support.
2. **When money comes in** — Fan pays tip or recurring sub. Stripe charges the fan; we **transfer to the artist’s Connect account** on our **weekly** schedule (same day for all artists). We do **not** hold artist funds in our platform balance or our bank.
3. **Payout to artist’s bank** — The artist’s Connect balance is paid out to their bank by **Stripe** on Stripe’s schedule and subject to Stripe’s minimum payout (if any). We don’t control that step; we never hold the money.

### Payout schedule and minimums

- **Our side** — We transfer to the artist’s Connect account **weekly**, on the **same day for all artists** (e.g. every Monday). That’s easy for us to forecast, and the regular drip of income shows artists the value and builds trust. We do not wait for a minimum balance before transferring; we never hold artist funds. If weekly isn’t feasible (e.g. Stripe or operational constraints), we fall back to **monthly**, same day for all.
- **Stripe’s side** — Stripe pays from the artist’s Connect balance to their bank on Stripe’s payout schedule. Any **minimum payout** (e.g. $10 or $25 before Stripe sends to the artist’s bank) is configured in Stripe and applies between Connect balance and bank. Artists see this in their Stripe dashboard; we document it in our help/terms.
- **Why** — This avoids us holding artist money, so we have no trust-money or interest-on-held-funds implications. Least legal exposure for us.

### Interest on held funds

We **do not hold** artist funds. We transfer to Connect immediately (or on a fixed schedule); Stripe holds the balance until payout to the artist’s bank. Stripe does not pay interest on Connect balances. So **no interest accrues to us**, and we have no obligation to pass interest to artists or to disclose interest. We never have artist money in our account.

### Fee handling (locked)

Stripe charges a fee for processing the fan’s payment (and possibly **for** payouts). We do **not** take a cut. We **deduct the fee from the artist’s payout**: fan pays $10; Stripe takes ~$0.59; artist receives ~$9.41. We are **upfront with artists** about handling fees — we show “$10 tip, $9.41 after fees” (or equivalent) in their dashboard and in our help/terms, so there are no surprises.

### Currency and multi-currency

- **We price in USD** — Subscription tiers and displayed prices are in US dollars ($10 / $20 / $50). USD reads as multi-national and is common for SaaS and music platforms; you can be based in the UK (or anywhere) and still show USD.
- **Stripe handles the rest** — Fans can pay with local payment methods; Stripe converts to USD (or we settle in USD). Artist payouts via Connect can be in the artist’s currency depending on Stripe/Connect setup. We don’t need to “support” multiple display currencies at launch; we quote in USD and Stripe does conversion where needed.

---

## Open decisions

- Tier names (Base / Growth / Unlimited or other) and exact track counts (25 / 100 / unlimited) — prices locked at $10 / $20 / $50; confirm before launch. Higher-tier perks (photo storage, custom domain, analytics) to add later.
- Confirm Stripe supports weekly transfer to Connect (same day for all); fallback to monthly if not.

