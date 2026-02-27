# Moderation and trust & safety

Content policy, reporting, and how we handle violations. We operate with a small team (per [Vision](./VISION.md)); moderation is in scope for “community health” but we can’t run a 24/7 review floor. We aim to be **transparent** and **trusted by musicians**.

## Implementation checklist

- Content policy: human-made only; no illegal, harassment, hate speech, impersonation, porn
- Explicit flag and age-gating for music, posts; minors protected (no stranger DMs when we add DMs)
- Reporting: signed-in only; report content (type, URL, reason); triage and review
- Moderation log: audit trail of created/updated content for proactive and reactive review
- Takedowns: platform removal; warnings; immediate takedown; suspension
- Copyright/DMCA-style: notice-and-takedown; counter-notice; repeat infringer
- Appeals: right to appeal; form/email; different reviewer; timeline and outcome

---

## Scope

- **Content policy** — What’s allowed on artist pages (music, feed, photos, gigs, site content) and on user-facing areas (profiles, platform galleries). Human-made music only; no harassment, hate speech, or illegal content.
- **Reporting** — Who can report, what they can report (artist page, single item, another user), and how (in-app form, email). How we triage and review with a small team.
- **Takedowns** — Platform-initiated vs report-driven; warnings vs immediate removal; account/page suspension when needed.
- **Copyright / DMCA-style** — If we host content that can infringe copyright (music, images), we need a process for rights-holder takedowns and (where applicable) counter-notice and repeat-infringer handling.
- **Appeals** — Whether artists (or users) can appeal a takedown or ban, and how.

---

## What we moderate

| Surface | Content | Who posts | Notes |
|--------|--------|-----------|--------|
| **Artist page** | Music, feed posts, images, YouTube embeds, photos, gig text, bio, links | Artist (owner + invitees) | Main focus; artist is responsible for page content. |
| **User-submitted photos** | Photos submitted to artist gallery | Users | Artist approves before public; we may still act if artist approves violating content or is a bad actor. |
| **User profile / platform** | Profile, personal gallery, blog (if any) | Users | Same policy as artist content where applicable. |
| **Discovery** | Artist name, handle, genre, location | Artist / system | Mostly metadata; impersonation or policy-breaching names can be reported. |
| **Comments** | Comments on posts, tracks, or artist pages (if we add them) | Users | In scope when we add comments; design must support reporting and moderation, and protect minors (see below). |
| **DMs / private messages** | Direct messages (if we add them) | Users | In scope when we add them; design must prevent strangers from contacting minors (see below). |

Comments and DMs are in scope when we add them; we design them so minors don’t talk with predators (e.g. age-aware restrictions on who can contact whom, reporting, moderation).

---

## Content policy (principles)

These are the principles; full public wording TBD.

1. **Human-made music only** — No AI-generated music. We don’t host or promote it. Other content (art, photos) should be created by humans; we don’t require proof but we act if we find systematic AI fakery that undermines the platform.
2. **No illegal content** — Content that is illegal in the jurisdictions we operate in (e.g. UK, and where we have users) is not allowed. We remove it and may report to authorities where required.
3. **No harassment or abuse** — Targeting individuals or groups with abuse, threats, doxxing, or coordinated harassment is not allowed. This applies to artist page content, profile content, and (if we add them) comments or messages.
4. **No hate speech** — Content that incites or promotes hatred or violence against people on the basis of protected characteristics (e.g. race, religion, sexuality, gender identity) is not allowed. We align with laws in our operating jurisdictions; when in doubt we err on the side of safety and clarity.
5. **No impersonation** — Artists and users must not impersonate other people or entities in a misleading way (parody and clear artistic persona may be acceptable; we judge in context).
6. **Copyright and trademarks** — Uploaders must have the right to use and display the content they upload. We have a separate process for rights-holder takedowns (see below). We don’t allow misleading use of others’ trademarks (e.g. impersonating a band or brand). “Free to download” does not mean “free to upload without permission”; we still require uploaders to have rights and we respond to valid notices.

7. **A-political; no personal attacks or violence** — We allow differing political views and controversial-but-legal art. We draw the line at: **personal attacks**, **cancellation attempts** (coordinated harassment to silence or harm someone), and **calls to violence** against anyone. We remove that content and may suspend.

8. **No porn or sexual smut** — We don’t allow pornographic content in photos, blogs, or audio. That includes: pornographic images, audio that is primarily sexual (e.g. smut audio stories, moaning or sexual-only tracks with no musical/artistic purpose). Artistic nudity or mature themes in music (lyrics about sex, drugs, etc.) are allowed when marked appropriately; we use an “explicit” flag and age-gate so minors can’t stream or download it (see “Minors, explicit content, and child safety” below).

---

## Minors, explicit content, and child safety

We allow music and art about mature themes (drugs, sex, violence in lyrics, etc.) but we **restrict minors** from accessing explicit content and we **protect minors** from contact with strangers. That aligns with **COPPA** (US) and the **UK Online Safety Act** (UK OSA).

### Explicit content (music, art)

- **Explicit flag** — Artists can mark tracks, posts, or other content as “explicit” (mature themes: strong language, sex, drugs, violence, etc.). We don’t censor artistic expression; we label it.
- **Age-gating** — Users must confirm they are **18 or over** (or meet a minimum age we set, e.g. 16+) before they can **stream or download** explicit content. We don’t show explicit content to users who haven’t passed the gate. How we verify age (e.g. self-declaration, account birthdate, or stronger verification for high-risk features) is TBD; legal advice will confirm what COPPA and UK OSA require.
- **Discovery** — We don’t promote explicit content to users who haven’t passed the age gate; we may hide or blur it in listings until the user is age-verified.

### Protecting minors from predators (comments, DMs)

- **Comments and DMs** — If we add comments or direct messages, we design so **minors don’t talk with predators**. That means: (1) **no unsolicited DMs from strangers to under-18s** (e.g. only people the minor follows or has accepted can DM them; or we disable DMs for under-18 accounts by default); (2) **comments** are visible and reportable so abuse can be caught and removed; (3) we **don’t recommend stranger accounts to children** (per UK OSA). Exact design TBD when we add comments/DMs; legal advice will confirm UK OSA and COPPA expectations.
- **Age / account** — We need a way to know if an account is under 18 (e.g. birthdate at signup, or “under 18” flag). We use that to apply stricter defaults (no DMs from strangers, no explicit content, no recommending strangers). We don’t collect more than necessary; legal advice will confirm what’s required for COPPA (e.g. verifiable parental consent for under-13) and UK OSA (e.g. age assurance for under-18).

### COPPA (US)

- **Under-13** — If we have users under 13 in the US, COPPA applies: we need **verifiable parental consent** before collecting personal information from them, and we must limit what we collect and how we use it. Options: (1) **Don’t allow under-13** (terms say 13+; we don’t knowingly collect from under-13); (2) **Allow under-13 with parental consent** (more complex; we’d need a compliant flow). We decide when we define signup; legal advice will confirm.
- **Actual knowledge** — If we have actual knowledge we’re collecting from a child under 13, we must comply (parental consent or cease collection). So we need a clear policy and, if we allow under-13, a compliant flow.

### UK Online Safety Act (UK OSA)

- **Children’s harmful content** — UK OSA requires platforms to protect **under-18s** from certain harmful content (e.g. pornography, eating-disorder content, self-harm). We: (1) **age-gate explicit content** so under-18s don’t access it; (2) **don’t allow porn or sexual smut** (see content policy); (3) use **age assurance** where required (e.g. to block under-18s from explicit content or from being contacted by strangers). Ofcom guidance will clarify exact duties; we follow it.
- **Stranger contact** — UK OSA expects platforms to **prevent strangers from messaging children**. So when we add DMs, we restrict who can message under-18 accounts (e.g. no unsolicited DMs from strangers).

*Exact age-verification method, under-13 policy, and UK OSA implementation to be confirmed with legal advice.*

---

## Reporting

- **Who can report** — **Signed-in users only.** They must be logged in to submit a report. We do not allow anonymous or signed-out reporting; that keeps abuse down and gives us a traceable reporter if we need to follow up.
- **What can be reported** — **All content types:** comments (when we have them), posts (feed, blog), photographs (artist gallery, user-submitted, platform galleries), music (tracks, albums), artist pages or profiles, and any other user-generated content we host. Report flow captures: what (type + URL or clear identifier), where it appears, and why (category: e.g. copyright, harassment, hate speech, illegal, explicit/porn, other).
- **How** — In-app “Report” control where the content is shown (on comments, posts, photos, tracks, profiles, etc.), plus a fallback (e.g. email to support with link and reason). We log reports and triage.
- **Review** — Small team; we can’t guarantee instant response. We prioritise: (1) illegal or imminent harm, (2) copyright with a valid notice, (3) harassment/hate, (4) other. We set a target (e.g. “we aim to review within X business days”) and state it in a help page or terms.
- **Outcome** — After review: no action, warning to poster, content takedown, or (in serious or repeated cases) account or page suspension. We notify the reporter where appropriate (e.g. “we’ve taken action”) without sharing private details; we notify the poster (see Takedowns).

---

## Moderation log (audit trail and proactive review)

We keep a **moderation log** of all user-generated content that is posted (or submitted for approval), so the team can review for policy violations as well as respond to reports.

- **What we log** — When content is created or updated: type (post, comment, photo, track, etc.), who posted it, when, where it appears (artist page, profile, etc.), and a reference to the content (e.g. ID, URL). We don’t necessarily store full content in the log forever (retention TBD); we need enough to triage and, when needed, retrieve the actual content for review.
- **How we use it** — The team can go through the log (e.g. chronological or filtered by type/surface) to look for material we don’t allow. This supports **proactive** moderation as well as **reactive** (reports). We may add **AI as a “first pass”** later: an automated system flags likely violations for human review; we do **not** auto-remove based on AI alone. Any such use is documented in our [AI use policy](./AI_USE_POLICY.md) as a permitted use (moderation first-pass detection).
- **Retention and access** — Log retention and access control (who can see the log, for how long) TBD; we align with privacy and data retention policy.

---

## Takedowns

- **Who can remove what** — **We** (platform) remove content that violates policy or that we must remove by law (e.g. court order, valid copyright notice). **Artists** can delete their own content anytime. **Users** can delete their own profile/content where the product allows.
- **Warnings** — For first-time or borderline cases we may warn and ask for edit/removal instead of immediate takedown. We keep a record (e.g. warning + date) for repeat-offender logic.
- **Immediate takedown** — For clear illegal content, serious harassment or hate speech, or valid copyright notice, we take down first and (except for copyright) notify the poster and explain appeal rights.
- **Suspension** — Repeated violations or one very serious violation can lead to temporary or permanent suspension of an artist page or user account. We document the reason and notify the user; appeal process applies.

---

## Copyright and trademarks: how we avoid infringement

Everything on the platform is **free to listen and download** for users. That does **not** mean anyone can **upload** whatever they like. We require uploaders to have the right to use and display the content; we respond to valid rights-holder notices; and we don’t allow misleading use of others’ trademarks.

### Copyright

- **Uploader responsibility** — By uploading, the artist or user represents that they have the **right** to use and display the content (e.g. they own it, have a licence, or have permission). We state this in our terms. We don’t pre-vet every upload for copyright; we rely on this representation and on a **notice-and-takedown** process.
- **Notice-and-takedown** — If a rights holder claims infringement, they send us a valid notice (see “Copyright and DMCA-style process” below). We remove or disable access to the content and notify the uploader. The uploader can send a counter-notice if they believe the notice was wrong; we may restore content after a set period unless the rights holder has filed court action. So: **free to download** = we don’t charge users for access; **we still enforce copyright** by taking down infringing uploads when we receive a valid notice.
- **Repeat infringer** — We may terminate accounts of users or artists who are repeatedly found to infringe, so we’re not hosting known bad actors.

### Trademarks

- **No misleading use** — We don’t allow uploaders to use others’ **trademarks** in a way that misleads (e.g. pretending to be another band or brand, or using a famous mark to imply endorsement). We remove or reject content that clearly infringes trademarks when we’re made aware (e.g. by report or notice).
- **Artist names and handles** — Artists choose their own names and handles. We don’t allow impersonation (see content policy). If someone reports that an artist name or handle infringes their trademark, we review and may require a change or removal.

### Summary

We avoid infringement by: (1) **terms** that require uploaders to have rights; (2) **notice-and-takedown** for copyright and trademark claims; (3) **repeat-infringer** termination; (4) **no impersonation** and no misleading use of others’ marks. “Free to download” is a **business choice** for users; it does **not** mean we permit uploads without permission.

---

## Copyright and DMCA-style process

We host user- and artist-uploaded music and images. Rights holders may claim infringement.

- **Liability** — We operate from the UK; we may have users and rightsholders in the US and elsewhere. We need a process that fits UK law and, if we target the US, something that aligns with DMCA safe-harbour style (notice, takedown, counter-notice, repeat infringer). Legal advice will confirm exact steps.
- **Rights-holder notice** — A clear way for a rights holder to submit a notice (form or email) with: work claimed, URL/location of infringing content, contact details, good-faith statement, and (where required) signature. We then remove or disable access to the content and notify the uploader.
- **Counter-notice** — If the uploader believes the notice was mistaken or they have rights, they can send a counter-notice. We restore content after a set period (e.g. 10–14 business days) unless the rights holder has filed court action. Again, exact process per jurisdiction and legal advice.
- **Repeat infringer** — We may terminate accounts of users or artists who are repeatedly found to infringe.

*Exact wording and flow to be defined with legal input; the above is the intended shape.*

---

## Appeals

- **Right to appeal** — Artists and users can appeal a takedown or suspension. We state this in the message we send when we take action.
- **How** — Reply to the notification email or use a designated “Appeal” path (e.g. form or email) with their reference (e.g. case ID) and short explanation. We don’t require legal language.
- **Who decides** — A different person than the one who took the action, where possible (small team may mean same person on different days; we still review with fresh eyes).
- **Timeline** — We aim to respond within a stated window (e.g. Y business days). Outcome: uphold, overturn (restore content / reinstate account), or modify (e.g. restore with warning).
- **Final** — We state that our appeal decision is final for the platform; users can still pursue legal routes if they believe their rights are violated.

---

## Transparency and trust

- **Public policy** — The full content policy (or a clear summary) is public on the site so artists and users know the rules.
- **Communication** — When we take action we explain in short form (e.g. “removed for [policy reason]”; “suspended for [reason]”) and point to appeal. We don’t share internal debate or reporter identity.
- **Stats (optional)** — We may publish high-level moderation stats (e.g. reports received, takedowns, appeals) in a yearly or periodic transparency note. TBD.

---

## Open decisions

- Full content policy wording (public-facing).
- Exact age-verification method for explicit content (self-declaration vs stronger; legal advice for COPPA and UK OSA).
- Under-13 policy: disallow (13+ only) vs allow with verifiable parental consent (legal advice).
- Target response times (review and appeal) and where we publish them.
- DMCA-style process: exact notice/counter-notice wording and repeat-infringer policy (with legal advice).
- Moderation log: retention period, access control, and whether to store full content or only references.
- Whether we add a “trusted reporter” or priority queue for repeat valid reports (abuse risk if not careful).
