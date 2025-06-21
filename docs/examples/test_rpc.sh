#!/bin/bash

# Test RPC Client for Basic Price Oracle
# This script runs the test RPC client to test the basic-price-oracle service

echo "🚀 Starting RPC Tests for Basic Price Oracle"
echo "=============================================="

# Check if the server is running
if ! pgrep -f "basic-price-oracle" > /dev/null; then
    echo "❌ Error: basic-price-oracle server is not running"
    echo "Please start the server first with: cd basic-price-oracle && ./basic-price-oracle"
    exit 1
fi

echo "✅ Server is running"

# Build the test client if it doesn't exist
if [ ! -f "./test_rpc_client" ]; then
    echo "🔨 Building test client..."
    go build test_rpc_client.go
    if [ $? -ne 0 ]; then
        echo "❌ Failed to build test client"
        exit 1
    fi
fi

echo "🧪 Running RPC tests..."
echo ""

# Run the test client
./test_rpc_client

echo ""
echo "✅ RPC tests completed!" 