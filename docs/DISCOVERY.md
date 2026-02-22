# Discovery

How users find artists: location, genre, search, and ranking.

---

## Artist location

Artists set their location using a **hierarchical** set of fields. They don’t have to fill all levels, but they go from broad to specific:

1. **Country** (required if they set any location)
2. **State / region** (e.g. California, Scotland — optional)
3. **County** (optional; not all regions use this)
4. **Town / city** (optional)

So someone might set only country; or country + state; or country + state + town. This lets users search for “artists near me” or “artists in Bristol” or “artists in Scotland.”

- Data model: we store whatever granularity the artist provides; search/filter respects the hierarchy (e.g. “artists in UK” includes everyone who set country=UK, regardless of whether they set town).
- We may need a consistent list of countries/regions/towns (e.g. from a standard dataset) so search is reliable — TBD.

---

## Genres

Artists choose a **limited number of genres** from a fixed list (e.g. 3–5). They can’t add custom genres; this keeps discovery consistent.

### Hierarchy: overarching genres and subgenres

Genres are organised as **overarching (parent) genres** with **subgenres**. An artist picks **one option per slot** — either a parent or a subgenre. Picking a **subgenre still counts as one slot**, but discovery treats it as belonging to the parent too.

- **Example:** If an artist picks **Folk Punk**, that uses **1 slot**. They can be found when a user browses or filters by **Punk** (the parent) or by **Folk Punk** (the subgenre). So “punk” discovery includes everyone who chose Punk or any subgenre of Punk (e.g. Folk Punk, Post-punk, Hardcore); “folk punk” narrows to only those who chose that subgenre.
- **Data model:** We store the artist’s choice(s) (e.g. “Folk Punk”). For discovery we either (1) store both the chosen label and its parent and filter by either, or (2) derive parent from subgenre when building search/filter indices. Either way: one pick = one slot; discovery by parent or by subgenre.

**Cap:** Max number of genre picks per artist (e.g. 3 or 5) — TBD. Artists pick from this list when setting up their page (or editing profile).

### Draft genre list (to be refined)

Structure: **Parent → subgenres**. Artists can pick a parent (e.g. Punk) or a subgenre (e.g. Folk Punk); picking a subgenre still uses one slot and allows discovery via both parent and subgenre.

- **Punk** → Folk Punk, Post-punk, Hardcore, (others TBD)
- **Folk** → Americana, Singer-songwriter, (others TBD)
- **Rock** → Indie rock, Alternative, (others TBD)
- **Metal** → Heavy metal, Doom, Black metal, (others TBD)
- **Pop** → Indie pop, (others TBD)
- **Electronic** → Techno, House, Ambient, IDM, (others TBD)
- **Hip-hop** → Rap, (others TBD)
- **Jazz** → Free jazz, Fusion, (others TBD)
- **Blues / Soul / R&B** → (subgenres TBD; may stay as broad parents)
- **Country** → Bluegrass, (others TBD)
- **Reggae** → Dub, Ska, (others TBD)
- **Classical** → Contemporary classical, (others TBD)
- **World / Traditional** → (subgenres TBD)
- **Experimental** → Noise, Drone, (others TBD)
- *(Other parents and subgenres TBD — final list should be manageable for both artists and search.)*

---

## Search and filter

Users can discover artists by:

- **Location** — e.g. “artists in my country,” “in this state,” “in this town,” or “near me” (if we have user location).
- **Genre** — one or more genres (AND/OR TBD; e.g. “punk AND folk” vs “punk OR folk”).
- **Combined** — e.g. “punk in Bristol,” “folk in Scotland.”
- **Free-text search** — Artist name, handle, and (optionally) bio. We support full-text search for artist names; this is a core part of discovery. Scope (name only vs name + bio + genre) TBD; all search will **probably use OpenSearch on AWS** (one search backend for filters, free-text, and relevance).

Search backend: **OpenSearch (AWS)** for indexing and querying artist data; exact index design and sync from primary store (e.g. DynamoDB) TBD. Artist and user lookups use a **discovery** index; feed and post search use a **separate feed index**. See [Architecture → OpenSearch: separate indices](./ARCHITECTURE.md#opensearch-separate-indices-for-discovery-vs-feed).

- **Block list** — For signed-in users, discovery (search, browse, charts) **excludes** artists and content from users they have blocked. Blocked artists don’t appear in results; content from blocked users (e.g. in any mixed feed) is hidden. See [Data and privacy → Blocking](./DATA_AND_PRIVACY.md#blocking-artists-and-other-users).

---

## Ranking / sort order

When showing search or browse results, we may offer:

- **Relevance** — from OpenSearch (full-text and filters).
- **Followers** — most followers first.
- **Last activity** — recently active (new post, new music, new gig) first.
- **Newest** — recently joined artists.
- **Alphabetical** — by name or handle.

Default sort TBD (e.g. last activity or followers). No paid boost or algorithmic engagement optimisation (per [Vision](./VISION.md)).

---

## Charts (download and listen)

We may offer **download** and **listen** charts for discoverability (e.g. “most downloaded this week,” “most played”). These are based only on **signed-in user** activity: we count listens and downloads when a signed-in user streams or downloads a track. Anonymous activity is not counted for charts, so charts reflect what logged-in users are actually playing and downloading. Exact definitions (e.g. listen = full play vs partial, time window for “this week”) and where charts appear (homepage, genre pages, etc.) TBD.

---

## Where discovery lives

- **Main site (www.afterwave.fm):** Homepage and browse/search experience — how this is structured (hero, “explore by location,” “explore by genre,” search bar) TBD.
- **Artist subdomains (handle.afterwave.fm):** Artist pages stay **artist-first**. We do **not** add “more artists like this” or “artists near here” links that send visitors away to discovery; that would compete with the artist for attention and feels unfair. Discovery lives on the main site; artist pages are the artist’s place.

---

## Open decisions

- Final genre list and max genres per artist.
- Whether we use a standard geography dataset (countries, states, towns) and how we handle user “near me.”
- Default sort and which sort options to expose.
- OpenSearch: exact index schema, which fields are searchable (name, handle, bio, genre), and how we sync from DynamoDB (or primary store) to OpenSearch.
- Charts: listen vs download definitions, time windows (e.g. weekly, monthly), and where they appear (homepage, genre, location).
