#!/bin/bash

# Oracle gRPC Test Examples using grpcurl
# This script provides comprehensive test examples for the basic-price-oracle service

echo "🔍 Oracle gRPC Test Examples using grpcurl"
echo "=========================================="

# Configuration
ORACLE_HOST="localhost:9095"
SERVICE_NAME="oraclerpc.PriceOracle"

# Supported assets from the oracle implementation
SUPPORTED_ASSET_ID="728926a632b6e262a3835261a36f8c85405b49a4d9e2976c169e154bddfdbe55"
SUPPORTED_GROUP_KEY="025841d1a2f05cb808e2c386da681f16bf45a783a077f5331f38c6411fb0ce506c"
UNSUPPORTED_ASSET_ID="1111111111111111111111111111111111111111111111111111111111111111"
BTC_ASSET_ID="0000000000000000000000000000000000000000000000000000000000000000"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_header() {
    echo -e "\n${BLUE}=== $1 ===${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_info() {
    echo -e "${YELLOW}ℹ️  $1${NC}"
}

# Check if grpcurl is installed
if ! command -v grpcurl &> /dev/null; then
    print_error "grpcurl is not installed. Please install it first:"
    echo "  macOS: brew install grpcurl"
    echo "  Linux: go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest"
    exit 1
fi

# Check if the oracle server is running
print_info "Checking if oracle server is running..."
if ! grpcurl -insecure -plaintext $ORACLE_HOST list &> /dev/null; then
    print_error "Oracle server is not running or not accessible at $ORACLE_HOST"
    print_info "Please start the oracle server first:"
    echo "  cd basic-price-oracle && ./oracle"
    exit 1
fi

print_success "Oracle server is running and accessible"

# List available services
print_header "Available gRPC Services"
grpcurl -insecure -plaintext $ORACLE_HOST list

# List available methods for the PriceOracle service
print_header "Available Methods for PriceOracle Service"
grpcurl -insecure -plaintext $ORACLE_HOST list $SERVICE_NAME

# Show method details
print_header "QueryAssetRates Method Details"
grpcurl -insecure -plaintext $ORACLE_HOST describe $SERVICE_NAME.QueryAssetRates

echo ""
print_header "Test 1: Basic PURCHASE Query (Supported Asset ID)"
echo "Testing a basic purchase query with the supported asset ID..."

grpcurl -insecure -plaintext -d '{
  "subject_asset": {
    "asset_id_str": "'$SUPPORTED_ASSET_ID'"
  },
  "subject_asset_max_amount": 1000,
  "payment_asset": {
    "asset_id_str": "'$BTC_ASSET_ID'"
  },
  "transaction_type": "PURCHASE"
}' $ORACLE_HOST $SERVICE_NAME.QueryAssetRates

echo ""
print_header "Test 2: Basic SALE Query (Supported Asset ID)"
echo "Testing a basic sale query with the supported asset ID..."

grpcurl -insecure -plaintext -d '{
  "subject_asset": {
    "asset_id_str": "'$SUPPORTED_ASSET_ID'"
  },
  "subject_asset_max_amount": 1000,
  "payment_asset": {
    "asset_id_str": "'$BTC_ASSET_ID'"
  },
  "transaction_type": "SALE"
}' $ORACLE_HOST $SERVICE_NAME.QueryAssetRates

echo ""
print_header "Test 3: Group Key Query (Supported Group Key)"
echo "Testing a query using the supported group key instead of asset ID..."

grpcurl -insecure -plaintext -d '{
  "subject_asset": {
    "group_key_str": "'$SUPPORTED_GROUP_KEY'"
  },
  "subject_asset_max_amount": 1000,
  "payment_asset": {
    "asset_id_str": "'$BTC_ASSET_ID'"
  },
  "transaction_type": "PURCHASE"
}' $ORACLE_HOST $SERVICE_NAME.QueryAssetRates

echo ""
print_header "Test 4: Large Amount Query (Should trigger 1-minute expiry)"
echo "Testing a query with a large amount that should trigger extended expiry..."

grpcurl -insecure -plaintext -d '{
  "subject_asset": {
    "asset_id_str": "'$SUPPORTED_ASSET_ID'"
  },
  "subject_asset_max_amount": 1000000,
  "payment_asset": {
    "asset_id_str": "'$BTC_ASSET_ID'"
  },
  "transaction_type": "PURCHASE"
}' $ORACLE_HOST $SERVICE_NAME.QueryAssetRates

echo ""
print_header "Test 5: With Asset Rates Hint (Accept Proposed Rates)"
echo "Testing a query with asset rates hint to accept proposed rates..."

grpcurl -insecure -plaintext -d '{
  "subject_asset": {
    "asset_id_str": "'$SUPPORTED_ASSET_ID'"
  },
  "subject_asset_max_amount": 1000,
  "payment_asset": {
    "asset_id_str": "'$BTC_ASSET_ID'"
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
}' $ORACLE_HOST $SERVICE_NAME.QueryAssetRates

echo ""
print_header "Test 6: Unsupported Asset (Should Return Error)"
echo "Testing a query with an unsupported asset ID..."

grpcurl -insecure -plaintext -d '{
  "subject_asset": {
    "asset_id_str": "'$UNSUPPORTED_ASSET_ID'"
  },
  "subject_asset_max_amount": 1000,
  "payment_asset": {
    "asset_id_str": "'$BTC_ASSET_ID'"
  },
  "transaction_type": "PURCHASE"
}' $ORACLE_HOST $SERVICE_NAME.QueryAssetRates

echo ""
print_header "Test 7: Missing Subject Asset (Should Return Error)"
echo "Testing a query without specifying the subject asset..."

grpcurl -insecure -plaintext -d '{
  "subject_asset_max_amount": 1000,
  "payment_asset": {
    "asset_id_str": "'$BTC_ASSET_ID'"
  },
  "transaction_type": "PURCHASE"
}' $ORACLE_HOST $SERVICE_NAME.QueryAssetRates

echo ""
print_header "Test 8: Non-BTC Payment Asset (Should Return Error)"
echo "Testing a query with a non-BTC payment asset..."

grpcurl -insecure -plaintext -d '{
  "subject_asset": {
    "asset_id_str": "'$SUPPORTED_ASSET_ID'"
  },
  "subject_asset_max_amount": 1000,
  "payment_asset": {
    "asset_id_str": "'$UNSUPPORTED_ASSET_ID'"
  },
  "transaction_type": "PURCHASE"
}' $ORACLE_HOST $SERVICE_NAME.QueryAssetRates

echo ""
print_header "Test 9: Different Amounts for Rate Comparison"
echo "Testing queries with different amounts to see rate variations..."

echo "Amount: 100"
grpcurl -insecure -plaintext -d '{
  "subject_asset": {
    "asset_id_str": "'$SUPPORTED_ASSET_ID'"
  },
  "subject_asset_max_amount": 100,
  "payment_asset": {
    "asset_id_str": "'$BTC_ASSET_ID'"
  },
  "transaction_type": "PURCHASE"
}' $ORACLE_HOST $SERVICE_NAME.QueryAssetRates

echo ""
echo "Amount: 10000"
grpcurl -insecure -plaintext -d '{
  "subject_asset": {
    "asset_id_str": "'$SUPPORTED_ASSET_ID'"
  },
  "subject_asset_max_amount": 10000,
  "payment_asset": {
    "asset_id_str": "'$BTC_ASSET_ID'"
  },
  "transaction_type": "PURCHASE"
}' $ORACLE_HOST $SERVICE_NAME.QueryAssetRates

echo ""
print_header "Test 10: USDF Asset (From NOTES.md)"
echo "Testing with the USDF asset ID mentioned in NOTES.md..."

USDF_ASSET_ID="5fd506e36846597e5699bdf550a20946a3af85bb2415aa4d74aad9e922d9053f"
USDF_GROUP_KEY="03b2bc6331eddcc5076eff6273718a46a37be274cee2b929e5dc5f666fcf3893c3"

echo "USDF Asset ID Query:"
grpcurl -insecure -plaintext -d '{
  "subject_asset": {
    "asset_id_str": "'$USDF_ASSET_ID'"
  },
  "subject_asset_max_amount": 1000,
  "payment_asset": {
    "asset_id_str": "'$BTC_ASSET_ID'"
  },
  "transaction_type": "PURCHASE"
}' $ORACLE_HOST $SERVICE_NAME.QueryAssetRates

echo ""
echo "USDF Group Key Query:"
grpcurl -insecure -plaintext -d '{
  "subject_asset": {
    "group_key_str": "'$USDF_GROUP_KEY'"
  },
  "subject_asset_max_amount": 1000,
  "payment_asset": {
    "asset_id_str": "'$BTC_ASSET_ID'"
  },
  "transaction_type": "PURCHASE"
}' $ORACLE_HOST $SERVICE_NAME.QueryAssetRates

echo ""
print_success "All grpcurl tests completed!"
print_info "Note: Some tests are expected to return errors to demonstrate error handling" 