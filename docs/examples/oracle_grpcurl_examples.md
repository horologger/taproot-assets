# Oracle gRPC Test Examples with grpcurl

This document provides individual grpcurl command examples for testing the basic-price-oracle service.

## Prerequisites

1. Install grpcurl:
   ```bash
   # macOS
   brew install grpcurl
   
   # Linux
   go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
   ```

2. Start the oracle server:
   ```bash
   cd basic-price-oracle
   ./oracle
   ```

## Configuration

- **Server Address**: `localhost:9095`
- **Service Name**: `oraclerpc.PriceOracle`
- **Method**: `QueryAssetRates`

## Asset IDs

- **Current Supported Asset ID**: `728926a632b6e262a3835261a36f8c85405b49a4d9e2976c169e154bddfdbe55`
- **Current Supported Group Key**: `025841d1a2f05cb808e2c386da681f16bf45a783a077f5331f38c6411fb0ce506c`
- **USDF Asset ID**: `5fd506e36846597e5699bdf550a20946a3af85bb2415aa4d74aad9e922d9053f`
- **USDF Group Key**: `03b2bc6331eddcc5076eff6273718a46a37be274cee2b929e5dc5f666fcf3893c3`
- **BTC Asset ID**: `0000000000000000000000000000000000000000000000000000000000000000`

## Basic Commands

### 1. List Available Services
```bash
grpcurl -insecure -plaintext localhost:9095 list
```

### 2. List Available Methods
```bash
grpcurl -insecure -plaintext localhost:9095 list oraclerpc.PriceOracle
```

### 3. Show Method Details
```bash
grpcurl -insecure -plaintext localhost:9095 describe oraclerpc.PriceOracle.QueryAssetRates
```

## Test Examples

### 1. Basic PURCHASE Query (Supported Asset ID)
```bash
grpcurl -insecure -plaintext -d '{
  "subject_asset": {
    "asset_id_str": "728926a632b6e262a3835261a36f8c85405b49a4d9e2976c169e154bddfdbe55"
  },
  "subject_asset_max_amount": 1000,
  "payment_asset": {
    "asset_id_str": "0000000000000000000000000000000000000000000000000000000000000000"
  },
  "transaction_type": "PURCHASE"
}' localhost:9095 oraclerpc.PriceOracle.QueryAssetRates
```

### 2. Basic SALE Query (Supported Asset ID)
```bash
grpcurl -insecure -plaintext -d '{
  "subject_asset": {
    "asset_id_str": "728926a632b6e262a3835261a36f8c85405b49a4d9e2976c169e154bddfdbe55"
  },
  "subject_asset_max_amount": 1000,
  "payment_asset": {
    "asset_id_str": "0000000000000000000000000000000000000000000000000000000000000000"
  },
  "transaction_type": "SALE"
}' localhost:9095 oraclerpc.PriceOracle.QueryAssetRates
```

### 3. Group Key Query (Supported Group Key)
```bash
grpcurl -insecure -plaintext -d '{
  "subject_asset": {
    "group_key_str": "025841d1a2f05cb808e2c386da681f16bf45a783a077f5331f38c6411fb0ce506c"
  },
  "subject_asset_max_amount": 1000,
  "payment_asset": {
    "asset_id_str": "0000000000000000000000000000000000000000000000000000000000000000"
  },
  "transaction_type": "PURCHASE"
}' localhost:9095 oraclerpc.PriceOracle.QueryAssetRates
```

### 4. Large Amount Query (Extended Expiry)
```bash
grpcurl -insecure -plaintext -d '{
  "subject_asset": {
    "asset_id_str": "728926a632b6e262a3835261a36f8c85405b49a4d9e2976c169e154bddfdbe55"
  },
  "subject_asset_max_amount": 1000000,
  "payment_asset": {
    "asset_id_str": "0000000000000000000000000000000000000000000000000000000000000000"
  },
  "transaction_type": "PURCHASE"
}' localhost:9095 oraclerpc.PriceOracle.QueryAssetRates
```

### 5. With Asset Rates Hint (Accept Proposed Rates)
```bash
grpcurl -insecure -plaintext -d '{
  "subject_asset": {
    "asset_id_str": "728926a632b6e262a3835261a36f8c85405b49a4d9e2976c169e154bddfdbe55"
  },
  "subject_asset_max_amount": 1000,
  "payment_asset": {
    "asset_id_str": "0000000000000000000000000000000000000000000000000000000000000000"
  },
  "transaction_type": "PURCHASE",
  "asset_rates_hint": {
    "subject_asset_rate": {
      "coefficient": "50000000000",
      "scale": 0
    },
    "payment_asset_rate": {
      "coefficient": "1000000000000",
      "scale": 0
    },
    "expiry_timestamp": 1750492008
  }
}' localhost:9095 oraclerpc.PriceOracle.QueryAssetRates
```

### 6. USDF Asset Query (From NOTES.md)
```bash
grpcurl -insecure -plaintext -d '{
  "subject_asset": {
    "asset_id_str": "5fd506e36846597e5699bdf550a20946a3af85bb2415aa4d74aad9e922d9053f"
  },
  "subject_asset_max_amount": 1000,
  "payment_asset": {
    "asset_id_str": "0000000000000000000000000000000000000000000000000000000000000000"
  },
  "transaction_type": "PURCHASE"
}' localhost:9095 oraclerpc.PriceOracle.QueryAssetRates
```

### 7. USDF Group Key Query
```bash
grpcurl -insecure -plaintext -d '{
  "subject_asset": {
    "group_key_str": "03b2bc6331eddcc5076eff6273718a46a37be274cee2b929e5dc5f666fcf3893c3"
  },
  "subject_asset_max_amount": 1000,
  "payment_asset": {
    "asset_id_str": "0000000000000000000000000000000000000000000000000000000000000000"
  },
  "transaction_type": "PURCHASE"
}' localhost:9095 oraclerpc.PriceOracle.QueryAssetRates
```

## Error Test Examples

### 8. Unsupported Asset (Should Return Error)
```bash
grpcurl -insecure -plaintext -d '{
  "subject_asset": {
    "asset_id_str": "1111111111111111111111111111111111111111111111111111111111111111"
  },
  "subject_asset_max_amount": 1000,
  "payment_asset": {
    "asset_id_str": "0000000000000000000000000000000000000000000000000000000000000000"
  },
  "transaction_type": "PURCHASE"
}' localhost:9095 oraclerpc.PriceOracle.QueryAssetRates
```

### 9. Missing Subject Asset (Should Return Error)
```bash
grpcurl -insecure -plaintext -d '{
  "subject_asset_max_amount": 1000,
  "payment_asset": {
    "asset_id_str": "0000000000000000000000000000000000000000000000000000000000000000"
  },
  "transaction_type": "PURCHASE"
}' localhost:9095 oraclerpc.PriceOracle.QueryAssetRates
```

### 10. Non-BTC Payment Asset (Should Return Error)
```bash
grpcurl -insecure -plaintext -d '{
  "subject_asset": {
    "asset_id_str": "728926a632b6e262a3835261a36f8c85405b49a4d9e2976c169e154bddfdbe55"
  },
  "subject_asset_max_amount": 1000,
  "payment_asset": {
    "asset_id_str": "1111111111111111111111111111111111111111111111111111111111111111"
  },
  "transaction_type": "PURCHASE"
}' localhost:9095 oraclerpc.PriceOracle.QueryAssetRates
```

## Compact One-Line Examples

For quick testing, here are compact one-line versions:

```bash
# Basic purchase query
grpcurl -insecure -plaintext -d '{"subject_asset":{"asset_id_str":"728926a632b6e262a3835261a36f8c85405b49a4d9e2976c169e154bddfdbe55"},"subject_asset_max_amount":1000,"payment_asset":{"asset_id_str":"0000000000000000000000000000000000000000000000000000000000000000"},"transaction_type":"PURCHASE"}' localhost:9095 oraclerpc.PriceOracle.QueryAssetRates

# Basic sale query
grpcurl -insecure -plaintext -d '{"subject_asset":{"asset_id_str":"728926a632b6e262a3835261a36f8c85405b49a4d9e2976c169e154bddfdbe55"},"subject_asset_max_amount":1000,"payment_asset":{"asset_id_str":"0000000000000000000000000000000000000000000000000000000000000000"},"transaction_type":"SALE"}' localhost:9095 oraclerpc.PriceOracle.QueryAssetRates

# USDF asset query
grpcurl -insecure -plaintext -d '{"subject_asset":{"asset_id_str":"5fd506e36846597e5699bdf550a20946a3af85bb2415aa4d74aad9e922d9053f"},"subject_asset_max_amount":1000,"payment_asset":{"asset_id_str":"0000000000000000000000000000000000000000000000000000000000000000"},"transaction_type":"PURCHASE"}' localhost:9095 oraclerpc.PriceOracle.QueryAssetRates
```

## Expected Responses

### Successful Response
```json
{
  "ok": {
    "assetRates": {
      "subjectAssetRate": {
        "coefficient": "102500000000",
        "scale": 0
      },
      "paymentAssetRate": {
        "coefficient": "1000000000000",
        "scale": 0
      },
      "expiryTimestamp": 1703123456
    }
  }
}
```

### Error Response
```json
{
  "error": {
    "message": "unsupported subject asset"
  }
}
```

## Notes

- The oracle server uses TLS with self-signed certificates, so we use `-insecure` flag
- The `-plaintext` flag is used for local testing (no TLS)
- Asset rates include a 5% spread (2.5% on each side of the base rate)
- Large amounts (>100,000) trigger extended expiry times
- Only BTC is supported as a payment asset
- The oracle supports both asset IDs and group keys for subject assets 