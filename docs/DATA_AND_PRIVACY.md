# Data, privacy, and export

Who owns what data; how we handle export, deletion, and privacy. We don’t sell data or use it for ad targeting (per [Vision](./VISION.md)).

---

## Scope

- **Notification / sign-up data** — Platform owns it; we don’t expose emails to artists.
- **User data** — Account, follows, downloads, payment history; what we store; what users can export or delete.
- **Artist data** — Artist page content, branding, music; artist owns their content; we host it. No raw subscriber emails to artists.
- **Privacy** — What we do and don’t do with data; GDPR and similar rights (access, rectify, delete, port).
- **Retention** — How long we keep data after account deletion or artist page removal.

---

## Notification sign-ups: we own it, we don’t expose emails to artists

When a user opts in to “be notified of new content on this artist page,” **we (the platform) own that data**. We store the link “user X is subscribed to notifications for artist Y” (and optionally whether they want email). We do **not** give artists the list of email addresses or any way to export or see who signed up.

- **Why** — Privacy: users signed up to get notifications from us about the artist, not to have their email handed to the artist. We send notifications on the artist’s behalf; the relationship is user ↔ platform ↔ artist.
- **What artists get** — Artists can see **aggregate counts** (e.g. “X people are subscribed to notifications”) and can **trigger a notification** (e.g. “notify my subscribers that I posted”) through our product. We send the notification; we don’t expose who received it or their contact details.
- **Export / portability** — We don’t export subscriber emails to artists. For **users**: they can see which artist pages they’re subscribed to and can export their own data (including “my notification subscriptions”) as part of their account export. So the user can take their list of subscriptions elsewhere; the artist never gets the underlying emails.
- **Email delivery** — If we support email notifications, we send them via our infrastructure (e.g. SES); recipients see the artist’s name and content, but the artist does not get bcc’d or given the recipient list.

See [Artist pages: Notifications](./ARTIST_PAGES_NOTIFICATIONS.md) for product behaviour; this doc is the data and privacy policy for that feature.

---

## User data

- **What we store** — Account (email, hashed password or OAuth id, profile if any), follows (which artists the user follows), notification subscriptions (which artist pages they’re subscribed to), **block list** (which artists and users they have blocked), downloads and listen history (for signed-in users, for charts and “recently played”), payment history (tips, artist subs, platform sub — for receipts and support). We don’t sell this; we don’t use it for ad targeting.
- **Export** — Users can **export their data**: profile, follows, notification subscriptions, **block list**, payment history (high-level: what they paid to whom, when), and optionally download/listen history. Format TBD (e.g. JSON or CSV); we provide it in a machine-readable way for portability.
- **Access and rectify** — Users can view and update their profile and preferences (including notification settings, follows, and block list) in the product. They can request a copy of their data (export) or correction; we support that in line with GDPR and similar laws.

### Blocking artists and other users

Users can **block artists** or **other users**. When they do, we store the block (e.g. “user A has blocked artist B,” “user A has blocked user C”) and **hide all content** from the blocked party for that user.

- **What is hidden** — For the blocker: we don’t show the blocked artist’s page, posts, music, or any of their content in discovery, search, charts, or feeds. We don’t show the blocked user’s profile, comments, posts, or any of their content. So the blocker never sees content from people they’ve blocked.
- **No notification** — The blocked artist or user is **not notified** that they’ve been blocked. We don’t show “who blocked you” to the blocked party.
- **Where it applies** — Discovery (search, browse, charts), artist pages (blocked artists’ pages are hidden or inaccessible to the blocker), comments (blocked users’ comments are hidden), feeds, and anywhere else we show artist or user content. We apply the block list whenever we render content for a signed-in user.
- **Reversible** — Users can unblock at any time from their account settings (e.g. “Blocked artists” / “Blocked users” list).
- **Data** — Block list is part of user data: included in export, cleared on account deletion. We don’t expose block lists to third parties or to the blocked party.

### Account deletion and sleep mode

When a user **deletes their account**, we remove their personal data (email, profile, follows, notification subscriptions) and they can choose what happens to their **comments** and **content they’ve uploaded**:

- **Delete all my comments** — We remove their comments (or replace with “deleted user” placeholder where threading requires it). User chooses yes/no.
- **Delete all content I’ve uploaded** — We remove content they uploaded (e.g. user-submitted photos, personal gallery, blog posts, any other user-generated content). User chooses yes/no.
- **Leave it behind** — If they choose **not** to delete comments or uploaded content, we **hide their identity** instead: their name and details are gone, but their comments and content **remain** and stay visible, attributed to “deleted user” or anonymous (so threads and galleries don’t break). Same outcome as sleep mode for visibility: name gone, details hidden, content remains.

**Sleep mode** — Users can put their account in **sleep mode** instead of deleting. Their **name is gone** and **details are hidden** (not visible to others); their **content remains** (comments, uploads stay visible but attributed to “deleted user” or anonymous). They can later **wake** the account (restore name and details) if we support it. Sleep mode is reversible; full deletion is not (unless we offer a grace period to undo).

We may retain some data where required by law (e.g. payment records for tax) in anonymised or aggregated form; see Retention. We don’t expose deleted or sleeping users’ personal data to artists or third parties.

---

## Artist data

- **Content** — Artists own their page content (music, posts, photos, bio, branding). We host it; they can delete it or leave the platform. We don’t claim ownership of their creative work; our terms grant us the licence we need to host and serve it.
- **Subscriber / notification data** — As above: we own it; we don’t expose emails or subscriber lists to artists. Artists get counts and the ability to trigger notifications through us only.
- **Export** — Artists can export **their content** (e.g. list of tracks, posts, photos — metadata and links to download their own files if we support it). They do **not** get an export of subscriber emails. They can see how many people are subscribed; they can’t see who.
- **Deletion** — Artists can delete their page or their content. When an artist page is removed, we stop sending notifications for that page and we remove the association “user X subscribed to artist Y” for that artist. We don’t hand the list to the artist on deletion.

---

## Privacy principles

- **No selling data** — We don’t sell user or artist data to third parties. Per Vision.
- **No ad targeting** — We don’t use user data for advertising or targeting. No ad-based revenue model.
- **Minimal use** — We use data to run the product: auth, notifications, discovery, payments, support, and compliance (e.g. tax records). We don’t use it for unrelated purposes.
- **GDPR and similar** — Where applicable (e.g. UK GDPR, EU, and other jurisdictions we operate in), we support: **access** (user can request a copy of their data), **rectification** (correct inaccurate data), **erasure** (delete account and personal data), **portability** (export in machine-readable form), and **object/restrict** where the law provides. We document the legal basis for processing (e.g. contract, consent) in our privacy policy. Privacy policy and terms of service: location and exact wording TBD; we own them and keep them up to date.

---

## Retention

- **Active accounts** — We retain user and artist data while the account or artist page is active. Users and artists can delete at any time.
- **After account deletion** — We remove or anonymise personal data (email, profile, follows, notification subscriptions). Per user choice: we remove their comments and/or uploaded content, or we leave content and hide their identity (name/details gone, content remains as “deleted user” or anonymous). We may retain **anonymised or aggregated** data for analytics or compliance. We retain **payment and tax records** as required by law (e.g. 7 years for tax; see [Tax and compliance](./TAX_AND_COMPLIANCE.md)) — these may reference the user by ID or transaction; we don’t use them for marketing or expose them to artists.
- **Sleep mode** — While the account is sleeping we hide name and details; content remains. If the user wakes the account we restore visibility; if they later delete we apply the same deletion choices (delete comments/content or leave behind with identity hidden).
- **After artist page removal** — We remove the artist’s content and the links “user X subscribed to artist Y” for that artist. We don’t retain a copy of the subscriber list for the artist.
- **Logs and backups** — Logs and backups may contain personal data for a limited period; we define retention for logs (e.g. 30–90 days) and ensure backups are overwritten or purged in line with deletion requests where feasible. Exact retention periods TBD and documented in privacy policy.

---

## Open decisions

- Exact export formats (JSON, CSV) and what fields are included for users and artists.
- Deletion: hard delete vs anonymise for each data type; grace period (e.g. 30 days to undo account deletion).
- Privacy policy and terms of service: final wording, URL, and legal review.
- Whether artists can “transfer” their notification audience (e.g. we send one final email “Artist X is now on Platform Y” with a link) — if we ever allow it, it would be user-initiated or with explicit user consent, not handing over emails to the artist.
