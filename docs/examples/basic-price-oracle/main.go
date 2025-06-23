// This example demonstrates a basic RPC price oracle server that implements the
// QueryAssetRates RPC method. The server listens on localhost:8095 and returns
// the asset rates for a given transaction type, subject asset, and payment
// asset.
// Add Kraken API websocket prices from.
// https://docs.kraken.com/rest/#operation/getAssetInfo
// https://docs.kraken.com/api/docs/websocket-v2/ticker
// https://docs.kraken.com/api/docs/guides/spot-ws-intro
// wss://ws.kraken.com/v2

package main

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"log"
	"math/big"
	"net"
	"os"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/lightninglabs/taproot-assets/rfqmath"
	"github.com/lightninglabs/taproot-assets/rfqmsg"
	oraclerpc "github.com/lightninglabs/taproot-assets/taprpc/priceoraclerpc"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	// serviceListenAddress is the listening address of the service.
	// serviceListenAddress = "localhost:8095"
	serviceListenAddress = "0.0.0.0:8095"

	// Kraken WebSocket configuration
	krakenWebSocketURL   = "wss://ws.kraken.com/v2"
	krakenReconnectDelay = 5 * time.Second
	krakenPingInterval   = 30 * time.Second
	krakenPongWait       = 60 * time.Second
	krakenWriteWait      = 10 * time.Second

	// supportedAssetIdStr is the hex-encoded asset ID for which this price
	// oracle provides exchange rates.
	// supportedAssetIdStr = "7b4336d33b019df9438e586f83c587ca00fa6560249" +
	// 	"7b93ace193e9ce53b1a67"

	// drewcoin
	supportedAssetIdStr = "11c6f5e7e84e9306c7ababacab239088f430fe14cab9c00c01ba6a9857cc4a70"

	// supportedGroupKeyStr is the hex-encoded asset group key for which
	// this price oracle provides exchange rates.
	// supportedGroupKeyStr = "02875ce409b587a6656357639d099ad9eb08396d0d" +
	// 	"fea8930a45e742c81d6fc782"

	// drewcoin
	supportedGroupKeyStr = "028dcdee288a9ece152a5d61ec07d8330c31928497e1f3dbb7e7125852d69dd12d"

	// baseAssetRate is the base rate for the asset in TAP asset units per BTC.
	// This represents the "fair" market price before applying spread.
	baseAssetRate = 100_000_000_000

	// spreadPercentage is the spread percentage applied to create buy/sell rates.
	// A value of 5 means 5% spread (2.5% on each side).
	// The spread is applied as: buy rate = base rate + (base rate * spread/2)
	//                           sell rate = base rate - (base rate * spread/2)
	spreadPercentage = 5
)

// KrakenWebSocketClient represents a WebSocket client for Kraken API
type KrakenWebSocketClient struct {
	conn        *websocket.Conn
	url         string
	done        chan struct{}
	mu          sync.RWMutex
	isConnected bool
	lastPrice   float64
	lastUpdate  time.Time
}

// KrakenTickerMessage represents a ticker message from Kraken WebSocket
type KrakenTickerMessage struct {
	Method string `json:"method"`
	Params struct {
		Channel string `json:"channel"`
		Symbol  string `json:"symbol"`
	} `json:"params"`
	Data []struct {
		Last float64 `json:"last"`
		Time int64   `json:"time"`
	} `json:"data"`
}

// KrakenSubscribeMessage represents a subscription message for Kraken WebSocket
type KrakenSubscribeMessage struct {
	Method string `json:"method"`
	Params struct {
		Channel string   `json:"channel"`
		Symbol  []string `json:"symbol"`
	} `json:"params"`
}

// setupLogger sets up the logger to write logs to a file.
func setupLogger() {
	// Create a log file.
	flags := os.O_CREATE | os.O_WRONLY | os.O_APPEND
	file, err := os.OpenFile("basic-price-oracle-example.log", flags, 0666)
	if err != nil {
		logrus.Fatalf("Failed to open log file: %v", err)
	}

	// Create a multi-writer to write to both stdout and the file.
	multiWriter := io.MultiWriter(os.Stdout, file)

	// Set the output of logrus to the multi-writer.
	logrus.SetOutput(multiWriter)

	// Set the log level (optional).
	logrus.SetLevel(logrus.DebugLevel)

	// Set the log format (optional).
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
}

// NewKrakenWebSocketClient creates a new Kraken WebSocket client
func NewKrakenWebSocketClient() *KrakenWebSocketClient {
	return &KrakenWebSocketClient{
		url:         krakenWebSocketURL,
		done:        make(chan struct{}),
		isConnected: false,
	}
}

// Connect establishes a WebSocket connection to Kraken
func (k *KrakenWebSocketClient) Connect() error {
	k.mu.Lock()
	defer k.mu.Unlock()

	if k.isConnected {
		return nil
	}

	logrus.Infof("Connecting to Kraken WebSocket: %s", k.url)

	conn, _, err := websocket.DefaultDialer.Dial(k.url, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to Kraken WebSocket: %v", err)
	}

	k.conn = conn
	k.isConnected = true

	// Set up connection parameters
	k.conn.SetPongHandler(func(string) error {
		k.conn.SetReadDeadline(time.Now().Add(krakenPongWait))
		return nil
	})

	logrus.Info("Successfully connected to Kraken WebSocket")
	return nil
}

// Subscribe subscribes to a ticker channel for a specific symbol
func (k *KrakenWebSocketClient) Subscribe(symbol string) error {
	k.mu.RLock()
	if !k.isConnected {
		k.mu.RUnlock()
		return fmt.Errorf("not connected to Kraken WebSocket")
	}
	k.mu.RUnlock()

	subscribeMsg := KrakenSubscribeMessage{
		Method: "subscribe",
	}
	subscribeMsg.Params.Channel = "ticker"
	subscribeMsg.Params.Symbol = []string{symbol}

	message, err := json.Marshal(subscribeMsg)
	if err != nil {
		return fmt.Errorf("failed to marshal subscribe message: %v", err)
	}

	k.mu.Lock()
	defer k.mu.Unlock()

	err = k.conn.WriteMessage(websocket.TextMessage, message)
	if err != nil {
		return fmt.Errorf("failed to send subscribe message: %v", err)
	}

	logrus.Infof("Subscribed to ticker for symbol: %s", symbol)
	return nil
}

// ReadMessages reads and processes incoming WebSocket messages
func (k *KrakenWebSocketClient) ReadMessages() {
	for {
		select {
		case <-k.done:
			return
		default:
			k.mu.RLock()
			if !k.isConnected {
				k.mu.RUnlock()
				return
			}
			k.mu.RUnlock()

			_, message, err := k.conn.ReadMessage()
			if err != nil {
				logrus.Errorf("Error reading WebSocket message: %v", err)
				k.reconnect()
				continue
			}

			k.processMessage(message)
		}
	}
}

// processMessage processes incoming WebSocket messages
func (k *KrakenWebSocketClient) processMessage(message []byte) {
	// Check if this is a heartbeat message and suppress logging
	if string(message) == `{"channel":"heartbeat"}` {
		return // Skip processing heartbeat messages entirely
	}

	// logrus.Infof("Received raw message: %s", string(message))

	var tickerMsg KrakenTickerMessage
	err := json.Unmarshal(message, &tickerMsg)
	if err != nil {
		logrus.Errorf("Failed to unmarshal ticker message: %v", err)
		logrus.Infof("Raw message that failed to parse: %s", string(message))
		return
	}

	// logrus.Infof("Parsed ticker message: Method=%s, Channel=%s, Symbol=%s, Data length=%d",
	// 	tickerMsg.Method, tickerMsg.Params.Channel, tickerMsg.Params.Symbol, len(tickerMsg.Data))

	if len(tickerMsg.Data) > 0 {
		// logrus.Infof("Ticker data: %+v", tickerMsg.Data[0])
		price := tickerMsg.Data[0].Last

		k.mu.Lock()
		k.lastPrice = price
		k.lastUpdate = time.Now()
		k.mu.Unlock()

		logrus.Infof("Received price update: %f", price)
	} else {
		logrus.Infof("No ticker data in message")
	}
}

// GetLastPrice returns the last received price
func (k *KrakenWebSocketClient) GetLastPrice() (float64, time.Time) {
	k.mu.RLock()
	defer k.mu.RUnlock()
	return k.lastPrice, k.lastUpdate
}

// reconnect attempts to reconnect to the WebSocket
func (k *KrakenWebSocketClient) reconnect() {
	logrus.Info("Attempting to reconnect to Kraken WebSocket...")

	k.mu.Lock()
	k.isConnected = false
	if k.conn != nil {
		k.conn.Close()
	}
	k.mu.Unlock()

	time.Sleep(krakenReconnectDelay)

	err := k.Connect()
	if err != nil {
		logrus.Errorf("Failed to reconnect: %v", err)
		return
	}

	// Resubscribe to the ticker
	err = k.Subscribe("BTC/USD")
	if err != nil {
		logrus.Errorf("Failed to resubscribe: %v", err)
	}
}

// Close closes the WebSocket connection
func (k *KrakenWebSocketClient) Close() {
	k.mu.Lock()
	defer k.mu.Unlock()

	close(k.done)

	if k.conn != nil {
		k.conn.Close()
	}

	k.isConnected = false
	logrus.Info("Kraken WebSocket connection closed")
}

// parseFloat is a helper function to parse float strings
func parseFloat(s string) (float64, error) {
	var f float64
	_, err := fmt.Sscanf(s, "%f", &f)
	return f, err
}

// RpcPriceOracleServer is a basic example RPC price oracle server.
type RpcPriceOracleServer struct {
	oraclerpc.UnimplementedPriceOracleServer
	wsClient *KrakenWebSocketClient
}

// Global WebSocket client instance
var globalWSClient *KrakenWebSocketClient

// isSupportedAssetID returns true if the given asset ID is supported by the
// price oracle, and false otherwise.
func isSupportedAssetID(rpcAssetSpec *oraclerpc.AssetSpecifier) bool {
	// Ensure that the subject asset is set.
	if rpcAssetSpec == nil {
		logrus.Info("Subject asset is not set (nil)")
		return false
	}

	supportedAssetIdBytes, err := hex.DecodeString(supportedAssetIdStr)
	if err != nil {
		fmt.Println("Error decoding supported asset hex string:", err)
		return false
	}

	// Check the subject asset bytes if set.
	subjectAssetIdBytes := rpcAssetSpec.GetAssetId()
	if len(subjectAssetIdBytes) > 0 {
		logrus.Infof("Subject asset ID bytes populated: %x",
			supportedAssetIdBytes)
		return bytes.Equal(supportedAssetIdBytes, subjectAssetIdBytes)
	}

	subjectAssetIdStr := rpcAssetSpec.GetAssetIdStr()
	if len(subjectAssetIdStr) > 0 {
		logrus.Infof("Subject asset ID str populated: %s",
			supportedAssetIdStr)
		return subjectAssetIdStr == supportedAssetIdStr
	}

	logrus.Infof("Subject asset ID not set")

	return false
}

// isSupportedAssetGroupKey returns true if the given asset group key is
// supported by the price oracle, and false otherwise.
func isSupportedAssetGroupKey(
	rpcAssetSpec *oraclerpc.AssetSpecifier) bool {

	// Ensure that the subject asset is not nil.
	if rpcAssetSpec == nil {
		logrus.Info("Subject asset is not set (nil)")
		return false
	}

	supportedGroupKeyBytes, err := hex.DecodeString(supportedGroupKeyStr)
	if err != nil {
		fmt.Println("Error decoding supported asset group key hex "+
			"string:", err)
		return false
	}

	// Check the subject asset group key bytes if set.
	subjectGroupKeyBytes := rpcAssetSpec.GetGroupKey()
	if len(subjectGroupKeyBytes) > 0 {
		logrus.Infof("Subject asset group key bytes populated: %x",
			supportedGroupKeyBytes)
		return bytes.Equal(supportedGroupKeyBytes, subjectGroupKeyBytes)
	}

	subjectGroupKeyStr := rpcAssetSpec.GetGroupKeyStr()
	if len(subjectGroupKeyStr) > 0 {
		logrus.Infof("Subject asset group key str populated: %s",
			supportedGroupKeyStr)
		return subjectGroupKeyStr == supportedGroupKeyStr
	}

	logrus.Infof("Subject asset group key not set")

	return false
}

// isSupportedAssetID returns true if the given subject asset is
// supported by the price oracle, and false otherwise.
func isSupportedSubjectAsset(subjectAsset *oraclerpc.AssetSpecifier) bool {
	return isSupportedAssetID(subjectAsset) ||
		isSupportedAssetGroupKey(subjectAsset)
}

// getPurchaseRate returns the buy (purchase) rate for the asset. The unit of
// the rate is the number of TAP asset units per BTC.
//
// Suppose our TAP asset is a USD stablecoin. To support liquidity and precision
// in rate conversion, we mint 1,000,000 TAP asset units per 1.00 USD. Wallet
// software must therefore display 1 USD for every 1,000,000 TAP asset units.
// In effect, wallet software must divide TAP asset units by 1,000,000 (= 10^6)
// to show the correct USD value. We call the exponent in this conversion factor
// (in this example it's 6) the `decimalDisplay` of the asset. The
// decimal display for a given asset is immutable and defined at minting.
//
// All rates returned by the price oracle service to tapd nodes are expressed as
// TAP asset units per BTC.
//
// Suppose the real-world price of 1 BTC is $100,000.00. To express this as a
// rate in TAP asset units per BTC, multiply by the decimal display conversion
// factor (10^decimalDisplay):
//
// realWorldPrice      = 100,000.00
// decimalDisplay      = 6
// tapAssetUnitsPerBtc = realWorldPrice * (10^decimalDisplay)
//
//	= 100,000.00 * 1,000,000
//	= 100,000,000,000
//
// Therefore, the price oracle should return the rate:
//
// rfqmath.NewBigIntFixedPoint(100_000_000_000, 0)
//
// ## When is the FixedPoint representation useful?
//
// Suppose another TAP asset is highly valuable, and the buy rate is
// 0.00001 TAP asset units per BTC. Spending 1 BTC would return 0.00001 TAP
// asset units of our valuable asset.
//
// To maintain integer-based communication, representing such small fractional
// rates directly is impractical due to floating-point precision issues.
// Instead, the price oracle returns a FixedPoint representation with integer
// components.
//
// A FixedPoint number `F` consists of a coefficient `C` and scale exponent `s`:
//
// F = C * (10^-s)
//
// Another way to express this is:
//
// F = C / (10^s)
//
// Now, we know F = 0.00001, and we are free to choose any integer pair (C, s)
// that satisfies this equation. One simple choice is to let C = 1 and s = 5,
// since 1 * 10^-5 = 0.00001.
//
// So the rate can be represented as:
//
// rfqmath.NewBigIntFixedPoint(1, 5)
func getPurchaseRate() rfqmath.BigIntFixedPoint {
	// Try to get real-time price from WebSocket
	if globalWSClient != nil {
		price, lastUpdate := globalWSClient.GetLastPrice()
		if price > 0 && time.Since(lastUpdate) < 5*time.Minute {
			// Use real-time price: convert USD price to TAP asset units per BTC
			// Assuming 1 TAP asset unit = $1 (decimal display = 6)
			realTimeRate := uint64(price * 1_000_000) // Convert USD to TAP asset units

			// Apply spread
			spreadAmount := (realTimeRate * spreadPercentage) / 100
			buyRate := realTimeRate + (spreadAmount / 2)

			logrus.Infof("Using real-time price: $%.2f -> %d TAP units/BTC (buy rate)",
				price, buyRate)
			return rfqmath.NewBigIntFixedPoint(buyRate, 0)
		}
	}

	// Fallback to base rate calculation
	spreadAmount := (baseAssetRate * spreadPercentage) / 100
	buyRate := baseAssetRate + (spreadAmount / 2)
	logrus.Infof("Using base rate: %d TAP units/BTC (buy rate)", buyRate)
	return rfqmath.NewBigIntFixedPoint(uint64(buyRate), 0)
}

// getSaleRate returns the sell/sale rate for the asset. The units of the
// rate is the number of TAP asset units per BTC.
//
// NOTE: see getPurchaseRate for more information.
func getSaleRate() rfqmath.BigIntFixedPoint {
	// Try to get real-time price from WebSocket
	if globalWSClient != nil {
		price, lastUpdate := globalWSClient.GetLastPrice()
		if price > 0 && time.Since(lastUpdate) < 5*time.Minute {
			// Use real-time price: convert USD price to TAP asset units per BTC
			// Assuming 1 TAP asset unit = $1 (decimal display = 6)
			realTimeRate := uint64(price * 1_000_000) // Convert USD to TAP asset units

			// Apply spread
			spreadAmount := (realTimeRate * spreadPercentage) / 100
			sellRate := realTimeRate - (spreadAmount / 2)

			logrus.Infof("Using real-time price: $%.2f -> %d TAP units/BTC (sell rate)",
				price, sellRate)
			return rfqmath.NewBigIntFixedPoint(sellRate, 0)
		}
	}

	// Fallback to base rate calculation
	spreadAmount := (baseAssetRate * spreadPercentage) / 100
	sellRate := baseAssetRate - (spreadAmount / 2)
	logrus.Infof("Using base rate: %d TAP units/BTC (sell rate)", sellRate)
	return rfqmath.NewBigIntFixedPoint(uint64(sellRate), 0)
}

// getAssetRates returns the asset rates for a given transaction type and
// subject asset max amount.
func getAssetRates(transactionType oraclerpc.TransactionType,
	subjectAssetMaxAmount uint64) (oraclerpc.AssetRates, error) {

	// Determine the rate based on the transaction type.
	var subjectAssetRate rfqmath.BigIntFixedPoint
	if transactionType == oraclerpc.TransactionType_PURCHASE {
		subjectAssetRate = getPurchaseRate()
	} else {
		subjectAssetRate = getSaleRate()
	}

	// Set the rate expiry to 5 minutes by default.
	expiry := time.Now().Add(5 * time.Minute).Unix()

	// If the subject asset max amount is greater than 100,000, set the rate
	// expiry to 1 minute.
	if subjectAssetMaxAmount > 100_000 {
		expiry = time.Now().Add(1 * time.Minute).Unix()
	}

	// Marshal subject asset rate to RPC format.
	rpcSubjectAssetToBtcRate, err := oraclerpc.MarshalBigIntFixedPoint(
		subjectAssetRate,
	)
	if err != nil {
		return oraclerpc.AssetRates{}, err
	}

	// Marshal payment asset rate to RPC format.
	rpcPaymentAssetToBtcRate, err := oraclerpc.MarshalBigIntFixedPoint(
		rfqmsg.MilliSatPerBtc,
	)
	if err != nil {
		return oraclerpc.AssetRates{}, err
	}

	return oraclerpc.AssetRates{
		SubjectAssetRate: rpcSubjectAssetToBtcRate,
		PaymentAssetRate: rpcPaymentAssetToBtcRate,
		ExpiryTimestamp:  uint64(expiry),
	}, nil
}

// QueryAssetRates queries the asset rates for a given transaction type, subject
// asset, and payment asset. An asset rate is the number of asset units per
// BTC.
//
// Example use case:
//
// Alice is trying to pay an invoice by spending an asset. Alice therefore
// requests that Bob (her asset channel counterparty) purchase the asset from
// her. Bob's payment, in BTC, will pay the invoice.
//
// Alice requests a bid quote from Bob. Her request includes an asset rates hint
// (ask). Alice obtains the asset rates hint by calling this endpoint. She sets:
// - `SubjectAsset` to the asset she is trying to sell.
// - `SubjectAssetMaxAmount` to the max channel asset outbound.
// - `PaymentAsset` to BTC.
// - `TransactionType` to SALE.
// - `AssetRateHint` to nil.
//
// Bob calls this endpoint to get the bid quote asset rates that he will send as
// a response to Alice's request. He sets:
// - `SubjectAsset` to the asset that Alice is trying to sell.
// - `SubjectAssetMaxAmount` to the value given in Alice's quote request.
// - `PaymentAsset` to BTC.
// - `TransactionType` to PURCHASE.
// - `AssetRateHint` to the value given in Alice's quote request.
func (p *RpcPriceOracleServer) QueryAssetRates(_ context.Context,
	req *oraclerpc.QueryAssetRatesRequest) (
	*oraclerpc.QueryAssetRatesResponse, error) {

	// Ensure that the payment asset is BTC. We only support BTC as the
	// payment asset in this example.
	if !oraclerpc.IsAssetBtc(req.PaymentAsset) {
		logrus.Infof("Payment asset is not BTC: %v", req.PaymentAsset)

		return &oraclerpc.QueryAssetRatesResponse{
			Result: &oraclerpc.QueryAssetRatesResponse_Error{
				Error: &oraclerpc.QueryAssetRatesErrResponse{
					Message: "unsupported payment asset, " +
						"only BTC is supported",
				},
			},
		}, nil
	}

	// Ensure that the subject asset is set.
	if req.SubjectAsset == nil {
		logrus.Info("Subject asset is not set")
		return nil, fmt.Errorf("subject asset is not set")
	}

	// Ensure that the subject asset is supported.
	if !isSupportedSubjectAsset(req.SubjectAsset) {
		logrus.Infof("Unsupported subject asset: %v\n",
			req.SubjectAsset)

		return &oraclerpc.QueryAssetRatesResponse{
			Result: &oraclerpc.QueryAssetRatesResponse_Error{
				Error: &oraclerpc.QueryAssetRatesErrResponse{
					Message: "unsupported subject asset",
				},
			},
		}, nil
	}

	// Determine which asset rate to return.
	var (
		assetRates oraclerpc.AssetRates
		err        error
	)

	if req.AssetRatesHint != nil {
		// If the asset rates hint is provided, return it as the asset
		// rate. In doing so, we effectively accept the asset rates
		// proposed by our peer.
		logrus.Info("Suggested asset to BTC rate provided, " +
			"returning rate as accepted rate")

		assetRates = oraclerpc.AssetRates{
			SubjectAssetRate: req.AssetRatesHint.SubjectAssetRate,
			PaymentAssetRate: req.AssetRatesHint.PaymentAssetRate,
			ExpiryTimestamp:  req.AssetRatesHint.ExpiryTimestamp,
		}
	} else {
		// If an asset rates hint is not provided, fetch asset rates
		// from our internal system.
		logrus.Info("Suggested asset to BTC rate not provided, " +
			"querying internal system for rate")

		assetRates, err = getAssetRates(
			req.TransactionType, req.SubjectAssetMaxAmount,
		)
		if err != nil {
			return nil, err
		}
	}

	logrus.Infof("QueryAssetRates returning rates (subject_asset_rate=%v, "+
		"payment_asset_rate=%v)", assetRates.SubjectAssetRate,
		assetRates.PaymentAssetRate)

	return &oraclerpc.QueryAssetRatesResponse{
		Result: &oraclerpc.QueryAssetRatesResponse_Ok{
			Ok: &oraclerpc.QueryAssetRatesOkResponse{
				AssetRates: &assetRates,
			},
		},
	}, nil
}

// startService starts the given RPC server and blocks until the server is
// shut down.
func startService(grpcServer *grpc.Server) error {
	serviceAddr := fmt.Sprintf("rfqrpc://%s", serviceListenAddress)
	logrus.Infof("Starting RPC price oracle service at address: %s\n",
		serviceAddr)

	server := RpcPriceOracleServer{}
	oraclerpc.RegisterPriceOracleServer(grpcServer, &server)
	grpcListener, err := net.Listen("tcp", serviceListenAddress)
	if err != nil {
		return fmt.Errorf("RPC server unable to listen on %s",
			serviceListenAddress)
	}
	return grpcServer.Serve(grpcListener)
}

// Generate a self-signed TLS certificate and private key.
func generateSelfSignedCert() (tls.Certificate, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return tls.Certificate{}, err
	}

	keyUsage := x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature
	extKeyUsage := []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"basic-price-oracle"},
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(24 * time.Hour), // Valid for 1 day

		KeyUsage:              keyUsage,
		ExtKeyUsage:           extKeyUsage,
		BasicConstraintsValid: true,
	}

	certDER, err := x509.CreateCertificate(
		rand.Reader, &template, &template, &privateKey.PublicKey,
		privateKey,
	)
	if err != nil {
		return tls.Certificate{}, err
	}

	privateKeyBits, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		return tls.Certificate{}, err
	}

	certPEM := pem.EncodeToMemory(
		&pem.Block{Type: "CERTIFICATE", Bytes: certDER},
	)
	keyPEM := pem.EncodeToMemory(
		&pem.Block{Type: "EC PRIVATE KEY", Bytes: privateKeyBits},
	)

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return tls.Certificate{}, err
	}

	return tlsCert, nil
}

func main() {
	setupLogger()

	// Initialize and start the Kraken WebSocket client
	logrus.Info("Initializing Kraken WebSocket client...")
	globalWSClient = NewKrakenWebSocketClient()

	// Connect to Kraken WebSocket
	err := globalWSClient.Connect()
	if err != nil {
		logrus.Warnf("Failed to connect to Kraken WebSocket: %v", err)
		logrus.Info("Continuing with base rate pricing...")
	} else {
		// Subscribe to BTC/USD ticker
		err = globalWSClient.Subscribe("BTC/USD")
		if err != nil {
			logrus.Warnf("Failed to subscribe to XBT/USD ticker: %v", err)
		} else {
			// Start reading messages in a goroutine
			go globalWSClient.ReadMessages()
			logrus.Info("Kraken WebSocket client started successfully")
		}
	}

	// Start the mock RPC price oracle service.
	//
	// Generate self-signed certificate. This allows us to use TLS for the
	// gRPC server.
	tlsCert, err := generateSelfSignedCert()
	if err != nil {
		log.Fatalf("Failed to generate TLS certificate: %v", err)
	}

	// Create the gRPC server with TLS
	transportCredentials := credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{tlsCert},
	})
	backendService := grpc.NewServer(grpc.Creds(transportCredentials))

	err = startService(backendService)
	if err != nil {
		log.Fatalf("Start service error: %v", err)
	}

	// Clean up WebSocket connection
	if globalWSClient != nil {
		globalWSClient.Close()
	}

	backendService.GracefulStop()
}
