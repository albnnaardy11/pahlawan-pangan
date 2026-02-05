---
sidebar_position: 1
---

# POS Integration (Pahlawan-Connect)

Pahlawan Pangan provides a "Zero-Click" automation suite to eliminate manual data entry for restaurant staff.

## Seamless Synchronization

The **Pahlawan-Connect** middleware integrates directly with Point-of-Sale (POS) systems (e.g., Moka, Majoo, Esensi).

### How it works:

1.  **Inventory Event**: When a shift ends or stock remains unsold, the POS system triggers a webhook to our `/v1/integration/pos/sync` endpoint.
2.  **AI Validation**: Pahlawan-AI validates the surplus items against historical data to ensure accuracy.
3.  **Automatic Listing**: The item is instantly published to the Pahlawan Marketplace with **Dynamic Pricing** enabled.

## API Authentication

Partners must use **Partner JWT Tokens** (RS256) for all synchronization requests.

```bash
# Example Sync Request
curl -X POST https://api.pahlawanpangan.org/v1/integration/pos/sync \
  -H "Authorization: Bearer <PARTNER_TOKEN>" \
  -d '{
    "provider_id": "RE-12345",
    "batch_items": [
      {"sku": "ROTI-01", "qty": 10, "expiry": "2024-12-31T21:00:00Z"}
    ]
  }'
```

## Benefits for Providers

- **Efficiency**: Staff save up to 45 minutes per day on manual reporting.
- **Accuracy**: Real-time inventory sync prevents "Ghost Items" (items sold in-store but still shown on app).
- **ROI Tracking**: Automatic generation of monthly sustainability and revenue recovery reports.
