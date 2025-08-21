# gRPC Reflection API Integration

This document explains the gRPC reflection API integration that has been added to the Basic Price Oracle server.

## What is gRPC Reflection?

gRPC reflection is a service that allows clients to discover the available services and methods on a gRPC server at runtime. This enables dynamic service discovery without requiring the client to have access to the `.proto` files.

## What was added?

### 1. Import Statement
```go
import (
    // ... existing imports ...
    "google.golang.org/grpc/reflection"
)
```

### 2. Reflection Service Registration
```go
// In main() function, after creating the gRPC server:
backendService := grpc.NewServer(grpc.Creds(transportCredentials))

// Register reflection service on gRPC server.
reflection.Register(backendService)
```

## Benefits of Reflection API

### 1. **Dynamic Service Discovery**
- Clients can discover available services without needing `.proto` files
- Tools like `grpcurl` can automatically discover the API structure
- Enables runtime introspection of the server

### 2. **Better Development Experience**
- IDEs can provide better autocomplete and documentation
- Debugging tools can show available methods and their signatures
- No need to manually specify service and method names

### 3. **Improved Testing**
- Test tools can automatically discover and test all available methods
- Easier integration testing without hardcoded service definitions

## Using Reflection with grpcurl

### List Available Services
```bash
grpcurl -insecure 127.0.0.1:9095 list
```

**Expected Output:**
```
grpc.reflection.v1alpha.ServerReflection
rfqrpc.RFQService
```

### List Methods for a Service
```bash
grpcurl -insecure 127.0.0.1:9095 list rfqrpc.RFQService
```

**Expected Output:**
```
rfqrpc.RFQService.QueryAssetRates
```

### Describe a Method
```bash
grpcurl -insecure 127.0.0.1:9095 describe rfqrpc.RFQService.QueryAssetRates
```

**Expected Output:**
```
rfqrpc.RFQService.QueryAssetRates is a method:
rpc QueryAssetRates ( .rfqrpc.QueryAssetRatesRequest ) returns ( .rfqrpc.QueryAssetRatesResponse );
```

### Describe Request/Response Types
```bash
# Describe the request type
grpcurl -insecure 127.0.0.1:9095 describe rfqrpc.QueryAssetRatesRequest

# Describe the response type
grpcurl -insecure 127.0.0.1:9095 describe rfqrpc.QueryAssetRatesResponse
```

## Testing Reflection API

### 1. Start the Server
```bash
cd basic-price-oracle
./oracle
```

### 2. Test Service Discovery
```bash
# In another terminal
grpcurl -insecure 127.0.0.1:9095 list
```

### 3. Test Method Discovery
```bash
grpcurl -insecure 127.0.0.1:9095 list rfqrpc.RFQService
```

### 4. Test Method Description
```bash
grpcurl -insecure 127.0.0.1:9095 describe rfqrpc.RFQService.QueryAssetRates
```

## Enhanced CLI Usage with Reflection

With reflection enabled, you can now use grpcurl more effectively:

### Interactive Mode
```bash
# Start grpcurl in interactive mode
grpcurl -insecure -plaintext 127.0.0.1:9095
```

### Tab Completion
Many shells and IDEs can now provide tab completion for:
- Service names
- Method names
- Field names in requests

### Automatic Type Discovery
```bash
# grpcurl can now automatically discover the correct request format
grpcurl -insecure 127.0.0.1:9095 rfqrpc.RFQService/QueryAssetRates -d '{
  "subject_asset": {
    "asset_id_str": "5fd506e36846597e5699bdf550a20946a3af85bb2415aa4d74aad9e922d9053f"
  },
  "subject_asset_max_amount": 1000,
  "payment_asset": {
    "asset_id_str": "0000000000000000000000000000000000000000000000000000000000000000"
  },
  "transaction_type": "PURCHASE",
  "asset_rates_hint": null
}'
```

## Security Considerations

### 1. **Information Disclosure**
- Reflection API exposes service and method information
- Consider disabling in production if this information is sensitive
- Can be controlled via configuration flags

### 2. **Access Control**
- Reflection API is available to any client that can connect to the server
- Consider implementing authentication/authorization if needed
- Can be restricted to specific IP ranges or users

### 3. **Production Deployment**
```go
// Example: Conditional reflection registration
if os.Getenv("ENABLE_REFLECTION") == "true" {
    reflection.Register(backendService)
}
```

## Troubleshooting

### Reflection Not Working
1. **Check if reflection is registered:**
   ```bash
   grpcurl -insecure 127.0.0.1:9095 list
   ```
   Should show `grpc.reflection.v1alpha.ServerReflection`

2. **Verify server is running:**
   ```bash
   nc -zv 127.0.0.1 9095
   ```

3. **Check server logs:**
   Look for any reflection-related error messages

### Common Issues
1. **Import not found:** Ensure `google.golang.org/grpc/reflection` is imported
2. **Registration failed:** Check that `reflection.Register()` is called after server creation
3. **Client can't connect:** Verify TLS settings and network connectivity

## Performance Impact

### Minimal Overhead
- Reflection API adds minimal runtime overhead
- Only active when clients request reflection information
- No impact on normal RPC calls

### Memory Usage
- Small increase in memory usage for reflection metadata
- Negligible for most applications

## Integration with Existing Tools

### 1. **grpcurl**
- Enhanced tab completion
- Automatic service discovery
- Better error messages

### 2. **gRPC UI Tools**
- Tools like grpcui can automatically discover services
- Better debugging experience

### 3. **IDEs and Editors**
- VS Code, IntelliJ, etc. can provide better gRPC support
- Auto-completion and documentation

## Future Enhancements

### 1. **Conditional Reflection**
```go
// Add environment variable control
if os.Getenv("ENABLE_REFLECTION") == "true" {
    reflection.Register(backendService)
    logrus.Info("gRPC reflection enabled")
}
```

### 2. **Reflection Metrics**
```go
// Add metrics for reflection usage
// (requires metrics framework)
```

### 3. **Access Control**
```go
// Add authentication for reflection API
// (requires auth middleware)
```

## Conclusion

The gRPC reflection API integration provides significant benefits for development, testing, and debugging while adding minimal overhead. It enables better tooling support and makes the API more discoverable for clients and developers.

The implementation is simple and follows gRPC best practices, making it easy to maintain and extend in the future. 