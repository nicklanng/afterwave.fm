# Afterwave.fm – Financial Scenarios & Growth Planning Model

All numbers are monthly.  
Currency: USD.  
These are directional estimates for planning — not accounting-grade forecasts.

## Implementation checklist

- Assumptions and model documented (this doc)
- Use for planning and runway scenarios; update as pricing/costs change
- N/A — planning doc; no code implementation

---

# Core Assumptions

## Platform Model

- Artist subscription: **$10/month**
- Supporter tier: **$4/month**
- Free tier: **64kbps Opus**
- Supporter tier: **128kbps Opus**
- Free cap: **100MB/month**
- Paid average usage: **500MB/month**
- CDN bandwidth cost: **$0.06/GB**
- Cognito: ~$0.004635 per MAU
- Stripe: 2.9% + $0.30 per subscription
- Corp tax: 25%

## Fixed Operating Costs (monthly)


| Item                  | Cost   |
| --------------------- | ------ |
| Fargate               | $1,200 |
| DynamoDB              | $800   |
| S3                    | $200   |
| Monitoring/misc infra | $500   |
| Accountant            | $250   |
| Legal                 | $500   |
| Tools                 | $300   |
| Insurance             | $250   |
| Contingency           | $500   |


Fixed non-bandwidth base ≈ **$4,500**

---

# Scenario 1 – Early Stage

### 100 Artists

### 50,000 MAU

### 1% Supporter Conversion (500 supporters)

---

## Revenue


| Source                | Amount     |
| --------------------- | ---------- |
| Artists (100 × $10)   | $1,000     |
| Supporters (500 × $4) | $2,000     |
| **Total Revenue**     | **$3,000** |


---

## Bandwidth

Free users: 49,500 × 100MB = 4.95TB  
Paid users: 500 × 500MB = 0.25TB  
Total ≈ 5.2TB  

Cost:  
5,200GB × $0.06 ≈ **$312**

---

## Cognito

50,000 × 0.004635 ≈ **$232**

---

## Stripe

2.9% of $2,000 = $58  
Fixed (500 × $0.30) = $150  
Total ≈ **$208**

---

## Total Expenses

Infra:

- Bandwidth: $312
- Cognito: $232
- Base infra: $4,500

Stripe: $208  

Total ≈ **$5,252**

---

## Profit

Revenue: $3,000  
Expenses: $5,252  

**Net: -$2,252 (loss phase)**

---

## Optional Ads (5 impressions/month, $20 CPM, 70% fill)

49,500 × 5 = 247,500 impressions  
Revenue ≈ $3,465  

New total revenue ≈ $6,465  

Net ≈ **+$1,213**

Ads stabilize early growth.

---

# Scenario 2 – Growth Phase

### 1,000 Artists

### 250,000 MAU

### 1% Supporter Conversion (2,500 supporters)

---

## Revenue


| Source     | Amount      |
| ---------- | ----------- |
| Artists    | $10,000     |
| Supporters | $10,000     |
| **Total**  | **$20,000** |


---

## Bandwidth

Free: 247,500 × 100MB = 24.75TB  
Paid: 2,500 × 500MB = 1.25TB  
Total ≈ 26TB  

Cost ≈ $1,560

---

## Cognito

250,000 × 0.004635 ≈ $1,159

---

## Stripe

2.9% of $10k = $290  
Fixed: 2,500 × $0.30 = $750  
Total ≈ $1,040

---

## Total Expenses

Infra:

- Bandwidth: $1,560
- Cognito: $1,159
- Base infra: $4,500

Stripe: $1,040  

Total ≈ **$8,259**

---

## Profit

Revenue: $20,000  
Expenses: $8,259  

Pre-tax ≈ $11,741  
After 25% tax ≈ **$8,806**

---

## Optional Ads (5 impressions/month)

247,500 × 5 = 1,237,500 impressions  
Revenue ≈ $17,325  

Total revenue ≈ $37,325  
Net profit ≈ **$21k+**

---

# Scenario 3 – 1M MAU Target

### 1,000 Artists

### 1,000,000 MAU

### 1% Supporter Conversion (10,000 supporters)

---

## Revenue


| Source     | Amount      |
| ---------- | ----------- |
| Artists    | $10,000     |
| Supporters | $40,000     |
| **Total**  | **$50,000** |


---

## Bandwidth

Free: 990,000 × 100MB = 99TB  
Paid: 10,000 × 500MB = 5TB  
Total ≈ 104TB  

Cost ≈ $6,240

---

## Cognito

1,000,000 × 0.004635 ≈ $4,635

---

## Stripe

2.9% of $40k = $1,160  
Fixed: 10,000 × $0.30 = $3,000  
Total ≈ $4,160

---

## Total Expenses

Infra:

- Bandwidth: $6,240
- Cognito: $4,635
- Base infra: $4,500

Stripe: $4,160  

Total ≈ **$19,535**

---

## Profit

Pre-tax ≈ $30,465  
After tax ≈ **$22,849**

---

## Optional Light Ads (5 impressions/month)

990,000 × 5 = 4,950,000 impressions  
At $20 CPM, 70% fill:

≈ $69,300

Total revenue ≈ $119,300  
Net profit after tax ≈ **~$75k+**

---

# Scenario 4 – 1M MAU, 5% Supporter Conversion

### 50,000 supporters

Revenue:

- Supporters: $200,000
- Artists: $10,000
- Total: $210,000

Bandwidth increases slightly (heavier paid usage):
≈ $8,000

Stripe:
≈ $20,800

Total expenses ≈ $40–45k

Pre-tax profit ≈ $165k  
After tax ≈ **~$123k/month**

Ads become irrelevant at this point.

---

# Strategic Observations

## 1. Early stage requires runway or ads.

Below ~100k MAU you likely operate at a loss unless:

- Conversion >1%
- Or ads fill gap.

## 2. 250k MAU is break-even + stable.

## 3. 1M MAU with 1% conversion is healthy and sustainable.

## 4. Ads are optional but extremely powerful stabilizers.

## 5. Conversion rate matters more than bitrate tweaks.

---

# Growth Planning Takeaways

- Get to 100k MAU → survival phase.
- 250k MAU → operational stability.
- 1M MAU → strong platform.
- 5% conversion → dominant economics.

---

If desired, we can next:

- Add a “team hiring” layer (1–3 engineers)  
- Or convert this into a Google Sheets growth simulator.

