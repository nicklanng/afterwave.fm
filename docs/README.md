# Afterwave.fm — Documentation

Documentation for the Afterwave.fm application.

## Contents

### Core product

- **[Vision](./VISION.md)** — Product overview, positioning vs Spotify/Bandcamp, and platform manifesto
- **[Sign-up and auth](./SIGNUP_AND_AUTH.md)** — User accounts, artist pages, subscription trigger, access control, and roles
- **[Artist pages](./ARTIST_PAGES.md)** — Overview: site builder, feed, notifications, music, photo gallery, gig calendar, support, merch. This is the main product that drives artist subscriptions. Each component has a detailed design doc linked from the overview.
- **[Discovery](./DISCOVERY.md)** — How users find artists: location (country → state → county → town), genres, search/filter, ranking
- **[Platform subscription](./PLATFORM_SUBSCRIPTION.md)** — Optional fan subscription to support the platform
- **[Music player app](./PLAYER_APP.md)** — Download-focused player for phone and desktop
- **[Payments](./PAYMENTS.md)** — Artist sub fee, tips, recurring artist subs, payouts

### Cross-cutting

- **[Data and privacy](./DATA_AND_PRIVACY.md)** — Ownership, export, deletion, GDPR-style
- **[Architecture](./ARCHITECTURE.md)** — Surfaces, stack, auth, storage, subdomains, technical providers (Stripe, AWS, DynamoDB, S3, OpenTelemetry)
- **[Deployment](./DEPLOYMENT.md)** — Environments, CI/CD, infra (AWS, compute, DB, S3, DNS), secrets, monitoring, rollback
- **[Go style guide](./GO_STYLE.md)** — Layout, naming, config, logging, HTTP, errors, testing, dependencies, formatting
- **[Phasing](./PHASING.md)** — MVP, what ships when, dependencies (scaffold)
- **[Moderation](./MODERATION.md)** — Content policy, reporting, takedowns

### Policy and ops

- **[Tax and compliance](./TAX_AND_COMPLIANCE.md)** — Merchant of record: multi-national tax (VAT/GST, sales tax), what we keep, auditing obligations
- **[AI use policy](./AI_USE_POLICY.md)** — How we use (and don’t use) AI in development and product

### Backlog

- **[Future improvements](./FUTURE_IMPROVEMENTS.md)** — Technical and product improvements we may do later (e.g. SQS-based OpenSearch indexing)
