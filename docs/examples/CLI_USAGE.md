# CLI Usage Guide - Testing with grpcurl

This guide demonstrates how to test the Basic Price Oracle API using the `grpcurl` command-line tool against a server running at `127.0.0.1:9095`.

## Prerequisites

1. **Install grpcurl**: Make sure you have `grpcurl` installed on your system
   ```bash
   # macOS
   brew install grpcurl
   
   # Ubuntu/Debian
   sudo apt-get install grpcurl
   
   # Or download from GitHub releases
   ```

2. **Server Running**: Ensure the Basic Price Oracle server is running at `127.0.0.1:9095`

## Basic Command Structure

```bash
grpcurl -insecure -d '{"request_json"}' 127.0.0.1:9095 priceoraclerpc.PriceOracle/QueryAssetRates
```

## Supported Asset Configuration

The server supports the following USDF asset:

- **Asset ID**: `5fd506e36846597e5699bdf550a20946a3af85bb2415aa4d74aad9e922d9053f`
- **Group Key**: `03b2bc6331eddcc5076eff6273718a46a37be274cee2b929e5dc5f666fcf3893c3`

## Test Cases

### 1. Basic Purchase Query

Query for purchasing USDF with BTC:

```bash
grpcurl -insecure -d '{
  "subject_asset": {
    "asset_id_str": "5fd506e36846597e5699bdf550a20946a3af85bb2415aa4d74aad9e922d9053f"
  },
  "subject_asset_max_amount": 1000,
  "payment_asset": {
    "asset_id_str": "0000000000000000000000000000000000000000000000000000000000000000"
  },
  "transaction_type": "PURCHASE",
  "asset_rates_hint": null
}' 127.0.0.1:9095 priceoraclerpc.PriceOracle/QueryAssetRates
```

**Expected Response**:
```json
{
  "ok": {
    "assetRates": {
      "subjectAssetRate": {
        "coefficient": "50000000000",
        "scale": 0
      },
      "paymentAssetRate": {
        "coefficient": "1000000000000",
        "scale": 0
      },
      "expiryTimestamp": "1750492008"
    }
  }
}
```

### 2. Sale Transaction Query

Query for selling USDF for BTC:

```bash
grpcurl -insecure -d '{
  "subject_asset": {
    "asset_id_str": "5fd506e36846597e5699bdf550a20946a3af85bb2415aa4d74aad9e922d9053f"
  },
  "subject_asset_max_amount": 1000,
  "payment_asset": {
    "asset_id_str": "0000000000000000000000000000000000000000000000000000000000000000"
  },
  "transaction_type": "SALE",
  "asset_rates_hint": null
}' 127.0.0.1:9095 priceoraclerpc.PriceOracle/QueryAssetRates
```

### 3. Using Group Key Instead of Asset ID

Query using the group key:

```bash
grpcurl -insecure -d '{
  "subject_asset": {
    "group_key_str": "03b2bc6331eddcc5076eff6273718a46a37be274cee2b929e5dc5f666fcf3893c3"
  },
  "subject_asset_max_amount": 1000,
  "payment_asset": {
    "asset_id_str": "0000000000000000000000000000000000000000000000000000000000000000"
  },
  "transaction_type": "PURCHASE",
  "asset_rates_hint": null
}' 127.0.0.1:9095 priceoraclerpc.PriceOracle/QueryAssetRates
```

### 4. Query with Asset Rates Hint

Provide a hint for the expected rates:

```bash
grpcurl -insecure -d '{
  "subject_asset": {
    "asset_id_str": "5fd506e36846597e5699bdf550a20946a3af85bb2415aa4d74aad9e922d9053f"
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
}' 127.0.0.1:9095 priceoraclerpc.PriceOracle/QueryAssetRates
```

### 5. Large Amount Query (> 100,000)

Test with a large amount to trigger extended expiry:

```bash
grpcurl -insecure -d '{
  "subject_asset": {
    "asset_id_str": "5fd506e36846597e5699bdf550a20946a3af85bb2415aa4d74aad9e922d9053f"
  },
  "subject_asset_max_amount": 200000,
  "payment_asset": {
    "asset_id_str": "0000000000000000000000000000000000000000000000000000000000000000"
  },
  "transaction_type": "PURCHASE",
  "asset_rates_hint": null
}' 127.0.0.1:9095 priceoraclerpc.PriceOracle/QueryAssetRates
```

### 6. Unsupported Asset Test

Test with an unsupported asset ID:

```bash
grpcurl -insecure -d '{
  "subject_asset": {
    "asset_id_str": "0000000000000000000000000000000000000000000000000000000000000001"
  },
  "subject_asset_max_amount": 1000,
  "payment_asset": {
    "asset_id_str": "0000000000000000000000000000000000000000000000000000000000000000"
  },
  "transaction_type": "PURCHASE",
  "asset_rates_hint": null
}' 127.0.0.1:9095 priceoraclerpc.PriceOracle/QueryAssetRates
```

**Expected Error Response**:
```json
{
  "error": {
    "message": "unsupported subject asset"
  }
}
```

### 7. Unsupported Payment Asset Test

Test with a non-BTC payment asset:

```bash
grpcurl -insecure -d '{
  "subject_asset": {
    "asset_id_str": "5fd506e36846597e5699bdf550a20946a3af85bb2415aa4d74aad9e922d9053f"
  },
  "subject_asset_max_amount": 1000,
  "payment_asset": {
    "asset_id_str": "0000000000000000000000000000000000000000000000000000000000000001"
  },
  "transaction_type": "PURCHASE",
  "asset_rates_hint": null
}' 127.0.0.1:9095 priceoraclerpc.PriceOracle/QueryAssetRates
```

**Expected Error Response**:
```json
{
  "error": {
    "message": "unsupported payment asset, only BTC is supported"
  }
}
```

## One-Liner Commands

For quick testing, here are compact one-liner versions:

### Basic Purchase
```bash
grpcurl -insecure -d '{"subject_asset":{"asset_id_str":"5fd506e36846597e5699bdf550a20946a3af85bb2415aa4d74aad9e922d9053f"},"subject_asset_max_amount":1000,"payment_asset":{"asset_id_str":"0000000000000000000000000000000000000000000000000000000000000000"},"transaction_type":"PURCHASE","asset_rates_hint":null}' 127.0.0.1:9095 priceoraclerpc.PriceOracle/QueryAssetRates
```

### Sale Transaction
```bash
grpcurl -insecure -d '{"subject_asset":{"asset_id_str":"5fd506e36846597e5699bdf550a20946a3af85bb2415aa4d74aad9e922d9053f"},"subject_asset_max_amount":1000,"payment_asset":{"asset_id_str":"0000000000000000000000000000000000000000000000000000000000000000"},"transaction_type":"SALE","asset_rates_hint":null}' 127.0.0.1:9095 priceoraclerpc.PriceOracle/QueryAssetRates
```

### Group Key Query
```bash
grpcurl -insecure -d '{"subject_asset":{"group_key_str":"03b2bc6331eddcc5076eff6273718a46a37be274cee2b929e5dc5f666fcf3893c3"},"subject_asset_max_amount":1000,"payment_asset":{"asset_id_str":"0000000000000000000000000000000000000000000000000000000000000000"},"transaction_type":"PURCHASE","asset_rates_hint":null}' 127.0.0.1:9095 priceoraclerpc.PriceOracle/QueryAssetRates
```

## Response Format

### Success Response
```json
{
  "ok": {
    "assetRates": {
      "subjectAssetRate": {
        "coefficient": "50000000000",
        "scale": 0
      },
      "paymentAssetRate": {
        "coefficient": "1000000000000",
        "scale": 0
      },
      "expiryTimestamp": "1750492008"
    }
  }
}
```

### Error Response
```json
{
  "error": {
    "message": "error description"
  }
}
```

## Rate Calculation

The rates returned represent:

- **Subject Asset Rate**: Price of 1 USDF in BTC (sats)
- **Payment Asset Rate**: Price of 1 BTC in sats (always 1,000,000,000,000)
- **Scale**: Decimal places for the coefficient

For example, with `"coefficient": "50000000000"` and `"scale": 0`:
- 1 USDF = 50,000 sats = 0.0005 BTC

## Troubleshooting

### Connection Issues
```bash
# Test if server is reachable
nc -zv 127.0.0.1 9095

# Check if gRPC service is available
grpcurl -insecure 127.0.0.1:9095 list
```

### TLS Issues
If the server uses proper TLS certificates, remove the `-insecure` flag:
```bash
grpcurl -d '{"request"}' 127.0.0.1:9095 priceoraclerpc.PriceOracle/QueryAssetRates
```

### JSON Format Issues
Use a JSON validator to ensure proper formatting:
```bash
# Validate JSON before sending
echo '{"your": "json"}' | jq .
```

## Script Examples

### Bash Script for Multiple Tests
```bash
#!/bin/bash

SERVER="127.0.0.1:9095"
ASSET_ID="5fd506e36846597e5699bdf550a20946a3af85bb2415aa4d74aad9e922d9053f"
GROUP_KEY="03b2bc6331eddcc5076eff6273718a46a37be274cee2b929e5dc5f666fcf3893c3"

echo "Testing Basic Price Oracle API at $SERVER"
echo "=========================================="

echo -e "\n1. Testing Purchase Query..."
grpcurl -insecure -d "{\"subject_asset\":{\"asset_id_str\":\"$ASSET_ID\"},\"subject_asset_max_amount\":1000,\"payment_asset\":{\"asset_id_str\":\"0000000000000000000000000000000000000000000000000000000000000000\"},\"transaction_type\":\"PURCHASE\",\"asset_rates_hint\":null}" $SERVER priceoraclerpc.PriceOracle/QueryAssetRates

echo -e "\n2. Testing Sale Query..."
grpcurl -insecure -d "{\"subject_asset\":{\"asset_id_str\":\"$ASSET_ID\"},\"subject_asset_max_amount\":1000,\"payment_asset\":{\"asset_id_str\":\"0000000000000000000000000000000000000000000000000000000000000000\"},\"transaction_type\":\"SALE\",\"asset_rates_hint\":null}" $SERVER priceoraclerpc.PriceOracle/QueryAssetRates

echo -e "\n3. Testing Group Key Query..."
grpcurl -insecure -d "{\"subject_asset\":{\"group_key_str\":\"$GROUP_KEY\"},\"subject_asset_max_amount\":1000,\"payment_asset\":{\"asset_id_str\":\"0000000000000000000000000000000000000000000000000000000000000000\"},\"transaction_type\":\"PURCHASE\",\"asset_rates_hint\":null}" $SERVER priceoraclerpc.PriceOracle/QueryAssetRates

echo -e "\n4. Testing Unsupported Asset..."
grpcurl -insecure -d "{\"subject_asset\":{\"asset_id_str\":\"0000000000000000000000000000000000000000000000000000000000000001\"},\"subject_asset_max_amount\":1000,\"payment_asset\":{\"asset_id_str\":\"0000000000000000000000000000000000000000000000000000000000000000\"},\"transaction_type\":\"PURCHASE\",\"asset_rates_hint\":null}" $SERVER priceoraclerpc.PriceOracle/QueryAssetRates

echo -e "\nTests completed!"
```

Save this as `test_api.sh` and run:
```bash
chmod +x test_api.sh
./test_api.sh
```

## Environment Variables

For easier testing, you can set environment variables:

```bash
export ORACLE_SERVER="127.0.0.1:9095"
export USDF_ASSET_ID="5fd506e36846597e5699bdf550a20946a3af85bb2415aa4d74aad9e922d9053f"
export USDF_GROUP_KEY="03b2bc6331eddcc5076eff6273718a46a37be274cee2b929e5dc5f666fcf3893c3"
export BTC_ASSET_ID="0000000000000000000000000000000000000000000000000000000000000000"
```

Then use them in commands:
```bash
grpcurl -insecure -d "{\"subject_asset\":{\"asset_id_str\":\"$USDF_ASSET_ID\"},\"subject_asset_max_amount\":1000,\"payment_asset\":{\"asset_id_str\":\"$BTC_ASSET_ID\"},\"transaction_type\":\"PURCHASE\",\"asset_rates_hint\":null}" $ORACLE_SERVER priceoraclerpc.PriceOracle/QueryAssetRates
``` 