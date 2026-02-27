# Tax and compliance (merchant of record)

As **merchant of record** we are responsible for collecting and remitting tax on our own revenue (artist subscriptions, platform fan subscriptions) and for keeping records that support tax filings and audits. This doc explains what that means in practice and what to do. **It is not legal or tax advice** — we must get professional advice (accountant, tax advisor) for our specific company structure and jurisdictions.

See [Payments](./PAYMENTS.md) for how we collect revenue (Stripe). See [Vision](./VISION.md) for our stance on clear pricing and no hidden fees.

## Implementation checklist

- Incorporate (e.g. UK limited company); engage accountant
- Register for sales tax (UK VAT, EU OSS when applicable, US state when nexus)
- Enable Stripe Tax; add registration numbers; collect customer location at checkout
- File returns and remit on schedule; keep records 7 years
- Invoices/receipts where required
- Document retention policy; assign who prepares filings
- When Connect (artist payouts): confirm reporting and tax form requirements

---

## Plain English summary

- **We sell subscriptions** (artist page sub, platform fan sub). The customer’s card shows “Afterwave” (or our legal entity). We are the **merchant of record**: we are legally on the hook for **sales-type tax** (VAT, GST, US state sales tax, etc.) in the places where we have to collect it.
- **Sales-type tax** = tax on the *sale* (e.g. VAT in the EU/UK, GST in Australia, state sales tax in the US). Rules for **digital services** (like software subscriptions) often depend on **where the customer is**, not where we are. So we may have to register and collect tax in multiple countries once we have enough customers there.
- **Stripe can calculate and collect** the right amount at checkout (Stripe Tax). Stripe does **not** register us with tax authorities or file returns — we (or our accountant) do that. So we: (1) register where required, (2) turn on Stripe Tax and give it our registration details, (3) file returns and pay the tax on time using Stripe’s reports.
- **We also pay income tax** on our profit (corporation tax in the UK, or equivalent where we’re resident). That’s separate from sales tax: we keep proper books, report revenue and costs, and pay tax on profit. An accountant does the company accounts and tax return.
- **We keep records** (invoices, transaction data, tax filings) for several years so we can prove what we charged, what tax we collected, and that we paid it. If we’re audited, we need to show that.

---

## Running before incorporation (free / coupon stage)

We can **run without a company for a while** to avoid company fees, accountants, and lawyers until we're ready to charge.

- **Free for artists via coupon** — If artist subscriptions are free (e.g. invite-only or coupon codes), we're not collecting artist sub revenue. If we also don't take platform fan subscriptions yet, we have **no subscription revenue** → no sales tax to collect, no company income to report. That's the simplest "pre-company" phase: build the product, onboard artists for free, get feedback. No legal entity required for $0 revenue.
- **Caveat** — As soon as we **charge** for artist subs or platform fan sub, we're taking revenue. Then we need a legal entity (company) and proper tax setup (registration, Stripe Tax, filings). Tips to artists go to artists via Connect — we're not the merchant for that money — so tips alone don't create subscription revenue for us. But if we take any fee or subscription ourselves, we need to be set up.
- **Use the free stage to document** — While artists are free (coupon stage), we document what we'll need when we flip to paid: retention policy (7 years, where to store), Stripe Tax setup checklist, which jurisdictions we'll register in first (e.g. UK VAT), and who'll do the accounts (accountant research). So when we launch paid subscriptions, we're ready to incorporate, register, and turn on Stripe Tax without a scramble.

**Summary:** Run without a company while revenue is zero (free artists via coupon, no platform sub). When we're ready to charge, incorporate and do the tax setup; we'll have already documented the steps in this phase.

---

## What "merchant of record" means

- **We** sell the subscription to the customer. The charge appears as Afterwave (or our legal entity) on their statement.
- **We** are liable for:
  - **Sales tax / VAT / GST** — Tax on the sale, in the jurisdictions where we have to collect it (based on customer location, thresholds, or “nexus”). We collect it at checkout, send it to the tax authority, and file returns.
  - **Income tax** — Tax on our *profit* (revenue minus costs). Our company pays corporation tax (or equivalent) where it’s resident. We keep books and file an annual return; an accountant typically does this.
- Stripe (and **Stripe Tax**) can **calculate** and **collect** the right sales tax at checkout and give us reports. They do **not** register us with tax authorities or file returns. We remain the MoR and must register, file, and remit ourselves (or via an accountant).

---

## Sales tax in practice (VAT / GST / US state)

We sell **digital services** (subscriptions). Most countries tax these and use the **customer’s location** to decide which rate and which country gets the tax.

### UK (if we’re UK-based)

- **UK VAT** — If we’re over the UK VAT threshold (~£90k/year turnover), we register for VAT with HMRC. We charge VAT on UK customers and on some overseas sales (rules depend on B2B vs B2C). For **digital services to UK consumers**, we charge UK VAT. Stripe Tax can apply the rate; we file a VAT return (usually quarterly) and pay HMRC.
- **Below threshold** — If we’re under the threshold, we don’t have to register (but we can voluntarily). Many small SaaS companies register anyway so they can reclaim VAT on costs.

### EU (customers in the EU)

- **EU VAT on digital services** — When we sell to **consumers (B2C)** in the EU, we generally have to charge VAT at the **customer’s country rate** (e.g. 20% in Germany, 19% in France). There’s a **€10,000/year threshold** (total cross-border B2C digital sales into the EU): below that, we can charge our home country VAT; above that, we charge each EU country’s VAT.
- **One-Stop Shop (OSS)** — Instead of registering in every EU country, we can use the **EU OSS** scheme: we register once (e.g. in the UK or one EU country) and report all EU B2C digital sales in one return. We remit the right amount per country through that one return. Stripe Tax can calculate per-country VAT; we use those numbers in our OSS return.
- **VAT ID** — For **business customers (B2B)** in the EU, we often don’t charge VAT if they give us a valid VAT ID (reverse charge). Stripe Tax can handle this if we collect the VAT number and pass it through.

### US (customers in the US)

- **US state sales tax** — There’s no single US VAT; each **state** (and sometimes city) can tax sales. Many states tax **digital products or SaaS**. We may have to register in states where we have “nexus” (physical presence or, in many states, **economic nexus** — e.g. over $100k sales or 200 transactions in that state). Once we have nexus, we collect that state’s sales tax and file returns there. Stripe Tax can calculate state-by-state; we register in each state where we have nexus and file (often quarterly).
- **Practical start** — Many small SaaS companies start with UK/EU only and add US state registrations once they have meaningful US revenue. We decide with our accountant.

**UK-based company selling to US customers**

- **We are UK-based.** Company is incorporated and tax-resident in the UK. We pay UK corporation tax on profit; we don’t file US **federal** income tax just because we have US customers.
- **Do US customers pay sales tax?** — **Yes, in states where we have nexus.** When a US customer in (e.g.) California subscribes and we have economic nexus in California (e.g. over $100k California sales or 200+ transactions there), we **collect** that state’s sales tax at checkout. The customer pays the tax (we add it to the charge); we **remit** it to the state. So US customers in nexus states pay sales tax; we don’t keep it — we pass it to the state.
- **Do we declare to the US?** — We don’t declare **income** to the US federal government (we’re UK-resident). We **do** have to **register and file in individual US states** where we have nexus: we file **state sales tax returns** (often quarterly) in each of those states and pay the tax we collected. So we “declare” (file returns) to **those states**, not to the federal US. Stripe Tax gives us the numbers; we (or our accountant) register in each nexus state and file. An accountant can tell us when we’ve crossed nexus in which states.

### Rest of world

- **Other countries** (e.g. Australia GST, Canada GST/HST, Norway VAT, etc.) often have similar rules for digital services: tax at customer’s location, thresholds, registration. We add jurisdictions as we grow or as our accountant advises. Stripe Tax supports many countries; we still have to register and file where required.

---

## What we need to do (checklist)

1. **Company location** — We are **UK-based** (incorporated and tax-resident in the UK). That drives UK corporation tax and our first VAT/GST registration (UK VAT, then EU OSS if we have EU B2C sales).
2. **Register for sales tax** — Register with the tax authority in our home country (e.g. UK VAT with HMRC) and, if we use it, EU OSS. Add US state registrations when we have nexus there. An accountant can do the registration and tell us thresholds.
3. **Turn on Stripe Tax** — In Stripe, enable Stripe Tax for our products (subscriptions, digital services). Add our tax registration numbers (e.g. UK VAT number, OSS scheme ID). Stripe will then calculate and add the right tax at checkout and give us reports.
4. **Collect customer location** — Stripe needs to know where the customer is (billing address, IP, or both). We ensure we collect billing country (and state for US) at checkout so Stripe Tax can apply the right rate.
5. **File returns and remit** — On the schedule required by each jurisdiction (e.g. UK VAT quarterly, EU OSS quarterly, US state quarterly), we (or our accountant) use Stripe Tax reports (and our own records) to complete the return and pay the tax. Stripe doesn’t file for us.
6. **Invoices / receipts** — Where the law requires an invoice (e.g. EU B2B, or for our own records), we issue one with our VAT number, customer details, and tax breakdown. Stripe can provide receipt/invoice data; we can use Stripe’s templates or our own, and we keep copies for the retention period.
7. **Income tax** — We keep proper books (revenue, costs, payouts to artists are costs or pass-through depending on structure). We file an annual company tax return and pay corporation tax on profit. An accountant does this. We can use **Xero** for accounts; connect Stripe to Xero so transaction data flows in and we (or our accountant) can reconcile and prepare returns. See “Xero (accounts and tax)” below.

---

## Stripe Tax (what it does and doesn’t do)

- **Does:** Calculates the correct sales tax (VAT, GST, state tax) based on customer location and our product type (digital services). Adds the tax line at checkout. Stores the tax amount per transaction. Provides **reports** (e.g. “this month: this much VAT per country”) so we can fill in our VAT/OSS/state returns. Supports many countries and product types.
- **Does not:** Register us with any tax authority. File any return. Remit any tax. Tell us where we have to register (we or our accountant decide based on thresholds and nexus). So we still need to: register ourselves, file returns, pay the tax. Stripe Tax just makes calculation and reporting much easier.

---

## Xero (accounts and tax)

We can use **Xero** for accounts and tax preparation. It helps with UK and non-UK obligations and connects to Stripe.

- **Stripe ↔ Xero** — Stripe integrates with Xero. Transaction data (payments, fees, subscriptions) flows from Stripe into Xero, so we get automatic reconciliation: payments and fees line up with invoices/sales in Xero. We connect Stripe from within Xero (e.g. Settings → Payment Services) or via the Xero–Stripe app. That gives us one place for books: revenue, Stripe fees, and (when we add Connect) payouts to artists can all be recorded and reconciled in Xero.
- **Does Xero help with non-UK obligations?** — **Yes, for recording and preparing returns.** Xero supports multi-currency and multiple tax rates (VAT, GST, US state sales tax). We can set up tax rates per country (and per US state if we have nexus), and Xero will calculate and track tax on sales. It can **prepare** VAT/GST/sales tax returns in formats suitable for different jurisdictions (e.g. UK VAT return, EU OSS, US state). So for UK VAT, EU OSS, and US state sales tax, Xero helps with: (1) recording transactions and tax in the right buckets, (2) preparing the numbers for each return. It does **not** register us with any tax authority or file returns — we (or our accountant) still register and file in each jurisdiction. So Xero is the **books and tax-prep** layer; Stripe Tax gives us the **correct tax at checkout** and reports; Xero holds the reconciled data and helps us (or our accountant) produce the actual returns.
- **Summary** — Stripe for payments and tax calculation at checkout; Stripe Tax for reports; Xero for accounts, reconciliation, and preparing UK and multi-country tax returns. We still register and file (or our accountant does). Xero is a good fit for a UK-based company with UK, EU, and US sales.

---

## Records and retention

We keep records so we can prove what we sold, what tax we charged, and that we filed and paid. If a tax authority or auditor asks, we must be able to show this.

### What to keep

- **Transaction records** — Who paid what, when, for what (subscription type, amount, tax amount). Stripe holds this; we export and archive (e.g. CSV or Stripe’s reports) so we’re not dependent on Stripe alone.
- **Tax returns and remittances** — Copies of every VAT/OSS/state return we file and proof of payment (e.g. bank transfer or HMRC confirmation).
- **Invoices / receipts** — Where we issue them, keep copies in a readable format (PDF). Include our VAT number and customer details where required.
- **Payout records** — For Stripe Connect (artist payouts): what we paid to whom, when. Needed for our own accounts and for any reporting we must do (e.g. 1099-K in the US if we’re required to report). See “Payouts to artists” below.

### Retention period

- **Typical requirement:** **7 years** after the end of the tax year (or the transaction) is a common rule in the UK and many other places. Some jurisdictions require longer. We set a **policy: keep transaction and tax records for 7 years** (or longer if our accountant says so). Store them in a way that survives system changes (exported files, backup, or archive storage). Don’t rely only on live DB; export periodically.

### Who is responsible

- **Who prepares filings** — We or our accountant. Document it (e.g. “Accountant X does UK VAT and EU OSS; we provide Stripe reports and transaction exports”).
- **Who responds to tax authority letters** — Usually our accountant or we with accountant guidance. Don’t ignore post from HMRC or other authorities.
- **Who owns retention** — We do. Even if the accountant has copies, we keep our own archive.

---

## Payouts to artists (Connect)

When we pay artists via Stripe Connect (tips, recurring fan→artist subs), we’re **facilitating** payments to them; we’re not the merchant for that money. But we may have **reporting** obligations:

- **US** — If we’re a US company or have US presence, we may have to report payments to US artists (e.g. **1099-K** above a threshold). If we’re UK-only and paying UK artists, US rules may not apply; our accountant will say. If we do have to report, we may need to collect tax forms (e.g. W-9) from artists.
- **UK / other** — Similar rules may apply (e.g. reporting payments to contractors). We confirm with our accountant as we enable Connect.
- **Artists’ own tax** — Artists are responsible for their own income tax on what they receive. We don’t withhold tax unless the law requires it. We document what we pay them (for our records and for any reporting we must do).

We expand this section when we implement Connect and know our company structure and artist base.

---

## Audits and reviews

- **Tax authority** — Can ask for evidence of revenue, tax collected, and remittances. We must be able to show: what we charged, what tax we applied, and that we filed and paid. Good records (see above) are essential.
- **Financial / internal** — For our own accounts, investor due diligence, or bank requests, the same records (transactions, tax, payouts) support this.
- **Stripe** — Stripe may do KYC or compliance checks. We keep our Stripe account in good standing (accurate business details, no prohibited use).

---

## Summary

| Area | Our responsibility | Stripe’s role |
|------|--------------------|---------------|
| **Sales tax (VAT/GST/state)** | Register, file returns, remit; set up Stripe Tax with our registration details | Calculate and collect at checkout; provide reports |
| **Records** | Keep transaction and tax records for 7 years (or as advised); export and archive | Store payment data; we export and retain what we need |
| **Invoices** | Issue where required; keep copies | Templates and data; we configure and retain |
| **Income tax** | Keep books; file annual return; pay corporation tax | N/A (we use Stripe data for our books) |
| **Payouts to artists** | Report where required; collect tax forms if needed | Connect payouts; we determine reporting |

---

## Next steps (practical)

**When we’re ready to take paid subscriptions** (artist subs or platform fan sub):

1. **Incorporate and choose location** (e.g. UK limited company). Get an **accountant** (or tax advisor) who knows small SaaS and digital services.
2. **Register for sales tax** in our home country (e.g. UK VAT). If we’ll have EU B2C sales, register for **EU OSS** (or equivalent) when we’re ready. Accountant can do this.
3. **Enable Stripe Tax** for our subscription products; add our VAT/OSS (and later US state) registration numbers. Ensure we collect billing country (and state for US) at checkout.
4. **Document retention policy** — “We keep transaction and tax records for 7 years; exports stored in [where].” Assign who prepares filings and who responds to tax post.
5. **When we add Connect (artist payouts)** — Ask accountant what we must report and whether we need to collect tax forms from artists. Update this doc and our product accordingly.

**Until then (free / coupon stage):** Run without a company; document retention policy, Stripe Tax checklist, and accountant research so we’re ready when we flip to paid.

**This doc is a guide, not advice.** Laws and thresholds change; our situation (company location, customer mix) will evolve. We rely on our accountant and tax advisor for decisions.
