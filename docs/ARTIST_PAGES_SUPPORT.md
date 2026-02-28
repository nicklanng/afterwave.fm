# Artist pages: Support / income

How tips and recurring subscriptions appear on the artist page and in the artist dashboard. Money flow (Stripe Connect, fees, payouts) is in [Payments](./PAYMENTS.md). This doc is about **where** and **what** fans and artists see.

See [Artist pages](./ARTIST_PAGES.md) for context. This is part of the core product that drives artist subscriptions.

## Implementation checklist

- Support section on artist page: tip button (one-off, anonymous or attributed), subscribe button (recurring, signed-in). Artists may also receive $5/month from **Support-tier** platform subscribers who select them (see [Platform subscription](./PLATFORM_SUBSCRIPTION.md)).
- Merch: link out only; no checkout
- Artist dashboard: tips list (anonymous/attributed), recurring subscribers list, payouts summary/link
- User account: payment history (tips, artist subs, platform sub); current artist subscriptions list + cancel
- Stripe Connect for tips and recurring; fee handling and transparency (see Payments)

---

## On the artist page (fan view)

- **Support section** — A block or section on the artist page (e.g. “Support” or “Tip / Subscribe”) with:
  - **One-off tip** — Button or form: fan enters amount (or picks preset), pays. No sign-in required; tip shows as “anonymous” to the artist. If signed in, tip is attributed (artist sees who gave). Fee is deducted from artist’s payout; we’re upfront about this (see Payments).
  - **Recurring subscription to artist** — Button: “Subscribe to [artist]” (or similar). Requires signed-in user. Fan pays periodically (e.g. monthly); 100% to artist minus processor fee. Same fee handling as tips.
- **Placement** — Where this section lives (e.g. sidebar, below bio, dedicated “Support” tab) is part of the site builder / layout. We just need a clear place for “tip” and “subscribe to artist” so fans can support without hunting.
- **Merch** — For now: artists add links (e.g. “Buy merch”) that point to their own store. We don’t handle checkout. Optional “Merch” section or link block in site builder; no design beyond “link out.”

---

## Artist dashboard (what artists see)

- **Tips** — List of tips: amount, date, and either “Anonymous” or the supporter’s name (if they were signed in). Export or filter TBD later.
- **Recurring subscribers** — List of fans who subscribe to this artist: who, plan (e.g. monthly), amount, start date. Plus list of **Support-tier** platform subscribers who have selected this artist to receive their $5/month. So the artist knows who’s supporting and can manage (e.g. thank them, or see churn). Export TBD later.
- **Payouts** — Link or summary to “your payouts” (Stripe Connect dashboard or our own summary). We don’t hold funds; payouts are weekly, same day for all (see Payments). Dashboard can show “next payout” and history.

---

## User account (fan view)

Users need a place to see **their** support activity and subscriptions. This lives in the main site (www.afterwave.fm when logged in), e.g. “Account” or “Billing” / “Support”.

- **Payment history** — List of what the user has paid: one-off tips (to whom, amount, date), recurring payments to artists, and (if applicable) platform fan subscription. So they can see where their money went and download receipts or export for their records. Scope: tips, artist subs, platform sub; we don’t need to show Stripe-level detail, just “you tipped X to [artist] on [date]” and “you pay [amount] to [artist] monthly.”
- **Current artist subscriptions** — List of artists the user is currently subscribed to (recurring). For each: artist name, amount, next charge date, and option to **cancel** or change. So users can manage their recurring support in one place instead of hunting on each artist page.

---

## Summary

| Who   | Sees                                                                 |
| ----- | -------------------------------------------------------------------- |
| Fan (on artist page) | Tip button (amount, one-off); Subscribe button (recurring, signed-in) |
| Fan (in account) | Payment history (tips, artist subs, platform sub); Current artist subscriptions (list + cancel) |
| Artist | Tips list (anonymous or attributed); Recurring subscribers list; Payout info |

Money flow, fees, Connect, and payouts: [Payments](./PAYMENTS.md). This doc: placement on page + dashboard view. No new product decisions; implementation detail only.
