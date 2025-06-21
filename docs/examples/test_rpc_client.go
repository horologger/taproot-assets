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

const (
	serverAddress = "localhost:8095"
)

func main() {
	// Create a TLS config that skips certificate verification for testing
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}

	// Create gRPC credentials
	creds := credentials.NewTLS(tlsConfig)

	// Connect to the server
	conn, err := grpc.Dial(serverAddress, grpc.WithTransportCredentials(creds))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// Create the client
	client := oraclerpc.NewPriceOracleClient(conn)

	// Test 1: Query asset rates for PURCHASE transaction
	fmt.Println("=== Test 1: PURCHASE Transaction ===")
	testPurchaseQuery(client)

	// Test 2: Query asset rates for SALE transaction
	fmt.Println("\n=== Test 2: SALE Transaction ===")
	testSaleQuery(client)

	// Test 3: Query with unsupported asset
	fmt.Println("\n=== Test 3: Unsupported Asset ===")
	testUnsupportedAsset(client)

	// Test 4: Query with asset rates hint
	fmt.Println("\n=== Test 4: With Asset Rates Hint ===")
	testWithAssetRatesHint(client)

	// Test 5: Query with large amount (should trigger 1-minute expiry)
	fmt.Println("\n=== Test 5: Large Amount Query ===")
	testLargeAmountQuery(client)
}

func testPurchaseQuery(client oraclerpc.PriceOracleClient) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create request for PURCHASE transaction
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
		log.Printf("Error querying asset rates: %v", err)
		return
	}

	printResponse(resp)
}

func testSaleQuery(client oraclerpc.PriceOracleClient) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create request for SALE transaction
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
		TransactionType: oraclerpc.TransactionType_SALE,
		AssetRatesHint:  nil,
	}

	resp, err := client.QueryAssetRates(ctx, req)
	if err != nil {
		log.Printf("Error querying asset rates: %v", err)
		return
	}

	printResponse(resp)
}

func testUnsupportedAsset(client oraclerpc.PriceOracleClient) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create request with unsupported asset
	req := &oraclerpc.QueryAssetRatesRequest{
		SubjectAsset: &oraclerpc.AssetSpecifier{
			Id: &oraclerpc.AssetSpecifier_AssetIdStr{
				AssetIdStr: "0000000000000000000000000000000000000000000000000000000000000001",
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
		log.Printf("Error querying asset rates: %v", err)
		return
	}

	printResponse(resp)
}

func testWithAssetRatesHint(client oraclerpc.PriceOracleClient) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create asset rates hint
	hint := &oraclerpc.AssetRates{
		SubjectAssetRate: &oraclerpc.FixedPoint{
			Coefficient: "50000000000",
			Scale:       0,
		},
		PaymentAssetRate: &oraclerpc.FixedPoint{
			Coefficient: "1000000000000",
			Scale:       0,
		},
		ExpiryTimestamp: uint64(time.Now().Add(10 * time.Minute).Unix()),
	}

	// Create request with asset rates hint
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
		AssetRatesHint:  hint,
	}

	resp, err := client.QueryAssetRates(ctx, req)
	if err != nil {
		log.Printf("Error querying asset rates: %v", err)
		return
	}

	printResponse(resp)
}

func testLargeAmountQuery(client oraclerpc.PriceOracleClient) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create request with large amount (> 100,000)
	req := &oraclerpc.QueryAssetRatesRequest{
		SubjectAsset: &oraclerpc.AssetSpecifier{
			Id: &oraclerpc.AssetSpecifier_AssetIdStr{
				AssetIdStr: "7b4336d33b019df9438e586f83c587ca00fa65602497b93ace193e9ce53b1a67",
			},
		},
		SubjectAssetMaxAmount: 200000, // > 100,000
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
		log.Printf("Error querying asset rates: %v", err)
		return
	}

	printResponse(resp)
}

func printResponse(resp *oraclerpc.QueryAssetRatesResponse) {
	switch result := resp.Result.(type) {
	case *oraclerpc.QueryAssetRatesResponse_Ok:
		assetRates := result.Ok.AssetRates
		fmt.Printf("✅ Success!\n")
		fmt.Printf("  Subject Asset Rate: %s (scale: %d)\n",
			assetRates.SubjectAssetRate.Coefficient, assetRates.SubjectAssetRate.Scale)
		fmt.Printf("  Payment Asset Rate: %s (scale: %d)\n",
			assetRates.PaymentAssetRate.Coefficient, assetRates.PaymentAssetRate.Scale)
		fmt.Printf("  Expiry Timestamp: %d (%s)\n",
			assetRates.ExpiryTimestamp,
			time.Unix(int64(assetRates.ExpiryTimestamp), 0).Format(time.RFC3339))

	case *oraclerpc.QueryAssetRatesResponse_Error:
		fmt.Printf("❌ Error: %s\n", result.Error.Message)
	}
}
