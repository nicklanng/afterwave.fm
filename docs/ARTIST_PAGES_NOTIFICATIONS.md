# Artist pages: Notifications

*Design doc for “be notified of new content on this artist page” — opt-in notifications, optional email.*

See [Artist pages](./ARTIST_PAGES.md) for context. This is part of the core product that drives artist subscriptions.

Data ownership and privacy (we own sign-up data; we don’t expose emails to artists) are in [Data and privacy](./DATA_AND_PRIVACY.md).

---

## Core idea

Fans **opt in** to “be notified of new content on this artist page.” The primary product is **in-app**: a **collated feed** of updates on the main page, plus a **bell** that shows a notification badge and opens the actual notification list. **Email delivery is optional** — we may offer it from day one or add it later; the platform sends emails on the artist’s behalf and does **not** give artists the list of subscriber emails.

---

## Opt-in flow

- **Where** — On the artist page: a clear control such as “Notify me” or “Get notified when [artist] posts.” Signed-in users only (we need a user to attach the subscription to).
- **Action** — One click (or one tap): user subscribes to notifications for that artist page. We record “user X is subscribed to notifications for artist Y.”
- **Optional: email** — If we support email, we can ask at opt-in: “Also email me when there’s new content?” (yes/no). Stored per user per artist. We do not expose email addresses to the artist; see [Data and privacy → Notification sign-ups](./DATA_AND_PRIVACY.md#notification-sign-ups-we-own-it-we-dont-expose-emails-to-artists).

---

## What triggers a notification

When the artist (or an invitee with the right permission) does certain actions, we **may** send a notification to everyone who opted in for that artist. Exact triggers to be decided; suggested:

| Trigger | Notify? | Notes |
| ------- | ------- | ----- |
| **New feed post** | Yes | Primary use case: “new post” is the main driver. |
| **New track or album** | Optional | “New music” — we can make this configurable (artist or subscriber preference) or always on. |
| **New gig** | Optional | “New gig added” — useful for fans who want to know when dates go up. |
| **Other** | TBD | e.g. new photo collection, major page update; lower priority for v1. |

We can start with “new feed post” only and add “new music” and “new gig” as options (per artist or per subscriber) later.

---

## What artists see and do

- **Counts** — Artists see **aggregate counts** only, e.g. “X people are subscribed to notifications.” They do **not** see who subscribed or any email addresses.
- **Trigger** — When the artist posts (or adds music/gigs), we send notifications on their behalf. The artist does not “upload a list” or “send an email themselves”; they use our product (e.g. “Publish post”) and we handle delivery. Optionally we expose “Notify subscribers” as an explicit action (e.g. for a post that’s already published) so they can send a one-off “new post” notification; TBD.

---

## Main page: collated feed

On the **main site** (e.g. www.afterwave.fm), signed-in users see a **collated feed** of feed updates from the artists they follow. This is the primary “what’s new” view: a single stream of new posts (and optionally new music, new gigs) from all followed artists, in one place. Chronological (newest first). So the user doesn’t have to visit each artist page to see new content — they get a combined activity feed on the main page.

This collated feed is **distinct** from the notification list (see Bell and notification badge below): the feed is the ongoing stream of updates; the bell is where we surface a **badge** and the list of notifications (e.g. unread).

---

## Bell and notification badge

- **Bell icon** — Notifications are accessed via a **bell** (e.g. in the main header/nav). The user has to **click the bell** to open the actual notification list. We do not show a notification count or badge anywhere else by default — the **badge lives on the bell**.
- **Badge** — When there are unread notifications, the bell shows a **badge** (e.g. a count or dot). Clicking the bell opens the notification list (e.g. dropdown or dedicated page); viewing the list (or individual items) can mark them read and update or clear the badge. Exact “read” behaviour TBD (e.g. mark-all-read, mark-on-view).
- **Notification list** — The list shows recent notifications: “[Artist] posted …”, “[Artist] added a new track …”, etc., with links to the relevant post or page. Same events that appear in the collated feed; the list is the “inbox” view, the feed is the stream view.

---

## What subscribers see (summary)

- **Main page** — Collated feed of updates from followed artists (see above).
- **Bell** — Badge when there are unread notifications; click bell to open notification list (see above).
- **Email** — If the subscriber opted in to email, we send the same (or a short summary) by email. From the platform (e.g. “Afterwave” or “Afterwave – [Artist name]”); artist does not see the recipient list.

---

## Unsubscribe

- **Per artist** — Subscribers can unsubscribe from an artist’s notifications at any time (e.g. from the artist page or from their notification settings). We remove the subscription; they stop receiving notifications for that artist.
- **Global** — Users can turn off all notification emails in account settings while keeping in-app notifications; TBD.

---

## Open decisions

- In-app only at first, or email from day one (see [Artist pages → Open decisions](./ARTIST_PAGES.md#open-decisions)).
- Exact list of triggers for v1 (post only vs post + music + gig).
- Whether artists can send a one-off “notify my subscribers” (e.g. for an existing post) or only automatic on publish.
- When the bell badge clears: mark-on-view, mark-all-read, or time-based (e.g. “read” after 7 days).
- Export for artists: we do not export emails; any “export” would be aggregate only (e.g. counts over time).
