# Price Oracle RPC Client Usage Guide

This document shows how to use the price oracle RPC client in your own applications.

## Basic Usage

Here's how to create and use a price oracle client in your Go application:

```go
package main

import (
    "context"
    "crypto/tls"
    "fmt"
    "log"
    "time"

    oraclerpc "github.com/lightninglabs/taproot-assets/taprpc/priceoraclerpc"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials"
)

func main() {
    // Create TLS config that skips certificate verification for testing
    tlsConfig := &tls.Config{
        InsecureSkipVerify: true,
    }

    // Create gRPC credentials
    creds := credentials.NewTLS(tlsConfig)

    // Connect to the server
    conn, err := grpc.Dial("localhost:8095", grpc.WithTransportCredentials(creds))
    if err != nil {
        log.Fatalf("Failed to connect: %v", err)
    }
    defer conn.Close()

    // Create the client
    client := oraclerpc.NewPriceOracleClient(conn)

    // Example: Get purchase rate
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    req := &oraclerpc.QueryAssetRatesRequest{
        SubjectAsset: &oraclerpc.AssetSpecifier{
            Id: &oraclerpc.AssetSpecifier_AssetIdStr{
                AssetIdStr: "7b4336d33b019df9438e586f83c587ca00fa65602497b93ace193e9ce53b1a67",
            },
        },
        SubjectAssetMaxAmount: 1000,
        PaymentAsset: &oraclerpc.AssetSpecifier{
            Id: &oraclerpc.AssetSpecifier_AssetIdStr{
                AssetIdStr: "0000000000000000000000000000000000000000000000000000000000000000",
            },
        },
        TransactionType: oraclerpc.TransactionType_PURCHASE,
        AssetRatesHint:  nil,
    }

    resp, err := client.QueryAssetRates(ctx, req)
    if err != nil {
        log.Fatalf("Error querying asset rates: %v", err)
    }

    switch result := resp.Result.(type) {
    case *oraclerpc.QueryAssetRatesResponse_Ok:
        assetRates := result.Ok.AssetRates
        fmt.Printf("Purchase Rate: %s (scale: %d)\n", 
            assetRates.SubjectAssetRate.Coefficient, 
            assetRates.SubjectAssetRate.Scale)
        fmt.Printf("Expires: %s\n", 
            time.Unix(int64(assetRates.ExpiryTimestamp), 0).Format(time.RFC3339))
        
    case *oraclerpc.QueryAssetRatesResponse_Error:
        fmt.Printf("Error: %s\n", result.Error.Message)
    }
}
```

## Key Concepts

### Asset Specification

Assets are specified using the `AssetSpecifier` message with a oneof field:

```go
// For hex string representation (recommended for REST)
assetSpec := &oraclerpc.AssetSpecifier{
    Id: &oraclerpc.AssetSpecifier_AssetIdStr{
        AssetIdStr: "7b4336d33b019df9438e586f83c587ca00fa65602497b93ace193e9ce53b1a67",
    },
}

// For raw bytes representation (gRPC only)
assetSpec := &oraclerpc.AssetSpecifier{
    Id: &oraclerpc.AssetSpecifier_AssetId{
        AssetId: []byte{...}, // 32-byte asset ID
    },
}
```

### BTC Asset

BTC is represented as an asset with all-zero asset ID:

```go
btcAsset := &oraclerpc.AssetSpecifier{
    Id: &oraclerpc.AssetSpecifier_AssetIdStr{
        AssetIdStr: "0000000000000000000000000000000000000000000000000000000000000000",
    },
}
```

### Transaction Types

```go
// For buying assets with BTC
transactionType := oraclerpc.TransactionType_PURCHASE

// For selling assets for BTC
transactionType := oraclerpc.TransactionType_SALE
```

### Rate Expiry

Rates have an expiry timestamp in Unix seconds:

```go
expiryTime := time.Unix(int64(assetRates.ExpiryTimestamp), 0)
if time.Now().After(expiryTime) {
    // Rate has expired, need to get a new one
}
```

## Common Patterns

### Getting Purchase Rate

```go
func getPurchaseRate(client oraclerpc.PriceOracleClient, amount uint64) (*oraclerpc.AssetRates, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    req := &oraclerpc.QueryAssetRatesRequest{
        SubjectAsset: &oraclerpc.AssetSpecifier{
            Id: &oraclerpc.AssetSpecifier_AssetIdStr{
                AssetIdStr: "7b4336d33b019df9438e586f83c587ca00fa65602497b93ace193e9ce53b1a67",
            },
        },
        SubjectAssetMaxAmount: amount,
        PaymentAsset: &oraclerpc.AssetSpecifier{
            Id: &oraclerpc.AssetSpecifier_AssetIdStr{
                AssetIdStr: "0000000000000000000000000000000000000000000000000000000000000000",
            },
        },
        TransactionType: oraclerpc.TransactionType_PURCHASE,
        AssetRatesHint:  nil,
    }

    resp, err := client.QueryAssetRates(ctx, req)
    if err != nil {
        return nil, fmt.Errorf("failed to query asset rates: %v", err)
    }

    switch result := resp.Result.(type) {
    case *oraclerpc.QueryAssetRatesResponse_Ok:
        return result.Ok.AssetRates, nil
    case *oraclerpc.QueryAssetRatesResponse_Error:
        return nil, fmt.Errorf("server error: %s", result.Error.Message)
    default:
        return nil, fmt.Errorf("unexpected response type")
    }
}
```

### Accepting Proposed Rates

```go
func acceptProposedRate(client oraclerpc.PriceOracleClient, proposedRate *oraclerpc.AssetRates, amount uint64) (*oraclerpc.AssetRates, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    req := &oraclerpc.QueryAssetRatesRequest{
        SubjectAsset: &oraclerpc.AssetSpecifier{
            Id: &oraclerpc.AssetSpecifier_AssetIdStr{
                AssetIdStr: "7b4336d33b019df9438e586f83c587ca00fa65602497b93ace193e9ce53b1a67",
            },
        },
        SubjectAssetMaxAmount: amount,
        PaymentAsset: &oraclerpc.AssetSpecifier{
            Id: &oraclerpc.AssetSpecifier_AssetIdStr{
                AssetIdStr: "0000000000000000000000000000000000000000000000000000000000000000",
            },
        },
        TransactionType: oraclerpc.TransactionType_PURCHASE,
        AssetRatesHint:  proposedRate,
    }

    resp, err := client.QueryAssetRates(ctx, req)
    if err != nil {
        return nil, fmt.Errorf("failed to query asset rates: %v", err)
    }

    switch result := resp.Result.(type) {
    case *oraclerpc.QueryAssetRatesResponse_Ok:
        return result.Ok.AssetRates, nil
    case *oraclerpc.QueryAssetRatesResponse_Error:
        return nil, fmt.Errorf("server error: %s", result.Error.Message)
    default:
        return nil, fmt.Errorf("unexpected response type")
    }
}
```

## Error Handling

The server can return various error conditions:

```go
switch result := resp.Result.(type) {
case *oraclerpc.QueryAssetRatesResponse_Ok:
    // Success - use result.Ok.AssetRates
    return result.Ok.AssetRates, nil
    
case *oraclerpc.QueryAssetRatesResponse_Error:
    // Server error - check result.Error.Message
    switch result.Error.Message {
    case "unsupported subject asset":
        // The asset is not supported by this oracle
        return nil, fmt.Errorf("asset not supported")
    case "unsupported payment asset, only BTC is supported":
        // Only BTC is supported as payment asset
        return nil, fmt.Errorf("only BTC payments supported")
    default:
        // Other server errors
        return nil, fmt.Errorf("server error: %s", result.Error.Message)
    }
    
default:
    return nil, fmt.Errorf("unexpected response type")
}
```

## Production Considerations

### TLS Configuration

For production use, you should use proper TLS certificates:

```go
// Load CA certificate
caCert, err := ioutil.ReadFile("ca.crt")
if err != nil {
    log.Fatalf("Failed to read CA cert: %v", err)
}

caCertPool := x509.NewCertPool()
caCertPool.AppendCertsFromPEM(caCert)

tlsConfig := &tls.Config{
    RootCAs: caCertPool,
}

creds := credentials.NewTLS(tlsConfig)
```

### Connection Management

```go
// Create connection with options
conn, err := grpc.Dial("localhost:8095", 
    grpc.WithTransportCredentials(creds),
    grpc.WithBlock(),
    grpc.WithTimeout(5*time.Second),
)
```

### Retry Logic

```go
func queryWithRetry(client oraclerpc.PriceOracleClient, req *oraclerpc.QueryAssetRatesRequest, maxRetries int) (*oraclerpc.AssetRates, error) {
    var lastErr error
    
    for i := 0; i < maxRetries; i++ {
        ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
        
        resp, err := client.QueryAssetRates(ctx, req)
        cancel()
        
        if err == nil {
            switch result := resp.Result.(type) {
            case *oraclerpc.QueryAssetRatesResponse_Ok:
                return result.Ok.AssetRates, nil
            case *oraclerpc.QueryAssetRatesResponse_Error:
                lastErr = fmt.Errorf("server error: %s", result.Error.Message)
                // Don't retry on server errors
                break
            }
        } else {
            lastErr = err
            // Wait before retry
            time.Sleep(time.Duration(i+1) * time.Second)
        }
    }
    
    return nil, fmt.Errorf("failed after %d retries: %v", maxRetries, lastErr)
}
```

## Testing

You can use the provided test client (`test_rpc_client.go`) as a reference for testing your implementation. The test client demonstrates all the common usage patterns and error conditions. 