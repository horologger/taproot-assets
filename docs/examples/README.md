# Basic Price Oracle RPC Test Suite

This directory contains a basic price oracle server implementation and comprehensive test suite for testing RPC queries.

## Overview

The basic price oracle server implements the `QueryAssetRates` RPC method from the Taproot Assets price oracle protocol. It provides asset exchange rates for TAP assets against BTC.

## Project Structure

```
.
├── basic-price-oracle/          # Main server implementation
│   ├── main.go                  # Server source code
│   ├── go.mod                   # Go module dependencies
│   └── basic-price-oracle       # Compiled server binary
├── test_rpc_client.go          # Go test client
├── test_rpc_client             # Compiled test client
├── test_rpc.sh                 # Shell script to run tests
├── go.mod                      # Go module dependencies for tests
└── README.md                   # This file
```

## Server Features

- **Supported Asset**: Only supports one specific asset ID (drewcoin)
- **Supported Group Key**: Also supports the asset group key for the same asset
- **Payment Asset**: Only supports BTC as payment asset
- **Transaction Types**: Supports both PURCHASE and SALE transactions
- **Rate Expiry**: Dynamic expiry based on transaction amount
- **TLS Security**: Uses self-signed certificates for secure communication

## Getting Started

### Prerequisites

- Go 1.23.9 (use `gvm use go1.23.9`)
- The basic-price-oracle server must be running

### Building the Server

```bash
# Build the server
make

# Or manually
cd basic-price-oracle
go build -v
```

### Starting the Server

```bash
cd basic-price-oracle
./basic-price-oracle
```

The server will start on `localhost:8095` with TLS enabled.

### Running the Tests

#### Option 1: Using the Shell Script (Recommended)

```bash
./test_rpc.sh
```

#### Option 2: Manual Testing

```bash
# Build the test client
go build test_rpc_client.go

# Run the tests
./test_rpc_client
```

## Test Cases

The test suite includes the following test scenarios:

### 1. PURCHASE Transaction
- Tests asset rates for buying the supported asset with BTC
- Expected: Returns purchase rate (102,500,000,000 asset units per BTC)

### 2. SALE Transaction  
- Tests asset rates for selling the supported asset for BTC
- Expected: Returns sale rate (97,500,000,000 asset units per BTC)

### 3. Unsupported Asset
- Tests query with an unsupported asset ID
- Expected: Returns error "unsupported subject asset"

### 4. Asset Rates Hint
- Tests the server's ability to accept proposed rates from peers
- Expected: Returns the provided hint rates instead of internal rates

### 5. Large Amount Query
- Tests rate expiry behavior for large transactions (>100,000 units)
- Expected: Returns 1-minute expiry instead of 5-minute expiry

### 6. Group Key Query
- Tests querying using the asset group key instead of asset ID
- Expected: Returns the same rates as asset ID queries

## API Reference

### QueryAssetRates RPC Method

**Request:**
```protobuf
message QueryAssetRatesRequest {
    TransactionType transaction_type = 1;
    AssetSpecifier subject_asset = 2;
    uint64 subject_asset_max_amount = 3;
    AssetSpecifier payment_asset = 4;
    uint64 payment_asset_max_amount = 5;
    AssetRates asset_rates_hint = 6;
}
```

**Response:**
```protobuf
message QueryAssetRatesResponse {
    oneof result {
        QueryAssetRatesOkResponse ok = 1;
        QueryAssetRatesErrResponse error = 2;
    }
}
```

### Asset Specification

Assets are specified using the `AssetSpecifier` message with a oneof field:

```protobuf
message AssetSpecifier {
    oneof id {
        bytes asset_id = 1;        // Raw bytes (gRPC)
        string asset_id_str = 2;   // Hex string (REST)
        bytes group_key = 3;       // Raw bytes (gRPC)
        string group_key_str = 4;  // Hex string (REST)
    }
}
```

### BTC Asset

BTC is represented as an asset with all-zero asset ID:
```
0000000000000000000000000000000000000000000000000000000000000000
```

### Supported Asset

The server supports only one specific asset (drewcoin):
```
Asset ID: 11c6f5e7e84e9306c7ababacab239088f430fe14cab9c00c01ba6a9857cc4a70
Group Key: 028dcdee288a9ece152a5d61ec07d8330c31928497e1f3dbb7e7125852d69dd12d
```

## Rate Calculation

The server uses a configurable spread-based pricing model:

### Configuration
- **Base Rate**: 100,000,000,000 asset units per BTC (represents the "fair" market price)
- **Spread Percentage**: 10% (configurable via `spreadPercentage` constant)
- **Spread Calculation**: The spread is split evenly between buy and sell rates

### Purchase Rate
- **Calculation**: Base rate + (Base rate × Spread percentage ÷ 2)
- **Current Rate**: 105,000,000,000 asset units per BTC
- **Meaning**: To buy 1 BTC worth of the asset, you need 105,000,000,000 asset units

### Sale Rate  
- **Calculation**: Base rate - (Base rate × Spread percentage ÷ 2)
- **Current Rate**: 95,000,000,000 asset units per BTC
- **Meaning**: When selling 95,000,000,000 asset units, you receive 1 BTC

### Adjusting the Spread
To change the spread, modify the `spreadPercentage` constant in `basic-price-oracle/main.go`:
```go
// 5% spread (2.5% on each side)
spreadPercentage = 5

// 10% spread (5% on each side)  
spreadPercentage = 10

// 20% spread (10% on each side)
spreadPercentage = 20
```

### Rate Expiry
- **Default**: 5 minutes for transactions ≤ 100,000 units
- **Large transactions**: 1 minute for transactions > 100,000 units

## Security

- **TLS**: Server uses self-signed certificates
- **Certificate Validity**: 24 hours from server start
- **Client**: Test client skips certificate verification for testing

## Troubleshooting

### Common Issues

1. **Server not running**
   ```
   ❌ Error: basic-price-oracle server is not running
   ```
   Solution: Start the server first with `cd basic-price-oracle && ./basic-price-oracle`

2. **Connection refused**
   ```
   Failed to connect: connection refused
   ```
   Solution: Check if server is running on port 8095

3. **TLS certificate errors**
   ```
   Failed to connect: tls: bad certificate
   ```
   Solution: The test client is configured to skip certificate verification

### Debugging

- Check server logs in `basic-price-oracle-example.log`
- Server runs on `localhost:8095` with TLS
- Use `ps aux | grep basic-price-oracle` to check if server is running

## Development

### Adding New Tests

To add new test cases, modify `test_rpc_client.go` and add new test functions following the existing pattern.

### Modifying Server Behavior

Edit `basic-price-oracle/main.go` to:
- Change supported asset IDs
- Modify rate calculations
- Adjust expiry times
- Add new validation logic

### Protocol Details

This implementation follows the Taproot Assets price oracle protocol specification. For more details, see the [Taproot Assets documentation](https://github.com/lightninglabs/taproot-assets). 