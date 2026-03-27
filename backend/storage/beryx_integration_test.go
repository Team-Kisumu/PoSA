// go:build integration

package storage

import (
	"os"
	"testing"
)

// Integration tests require BERYX_API_TOKEN in environment.
// Run with: go test -tags=integration -v ./storage/

func loadBeryxToken(t *testing.T) string {
	token := os.Getenv("BERYX_API_TOKEN")
	if token == "" {
		t.Skip("BERYX_API_TOKEN not set — skipping Beryx integration tests")
	}
	return token
}

// TestBeryxConnection verifies that the Beryx client can connect
// to the Filecoin network and retrieve the chain ID.
func TestBeryxConnection(t *testing.T) {
	token := loadBeryxToken(t)

	client, err := NewBeryxClient(BeryxConfig{
		RPCURL:   "https://api.zondax.ch/fil/node/mainnet/rpc/v1",
		APIToken: token,
	})
	if err != nil {
		t.Fatalf("NewBeryxClient failed: %v", err)
	}
	defer client.Close()

	chainID := client.ChainID()
	if chainID.Int64() != 314 {
		t.Errorf("chain ID = %d, want 314 (mainnet)", chainID.Int64())
	}
	t.Logf("Connected to Filecoin Mainnet (chain ID: %d)", chainID.Int64())
}

// TestBeryxBlockNumber verifies that the Beryx client can fetch
// the latest block number from the Filecoin network.
func TestBeryxBlockNumber(t *testing.T) {
	token := loadBeryxToken(t)

	client, err := NewBeryxClient(BeryxConfig{
		RPCURL:   "https://api.zondax.ch/fil/node/mainnet/rpc/v1",
		APIToken: token,
	})
	if err != nil {
		t.Fatalf("NewBeryxClient failed: %v", err)
	}
	defer client.Close()

	block, err := client.BlockNumber()
	if err != nil {
		t.Fatalf("BlockNumber failed: %v", err)
	}
	if block < 5000000 {
		t.Errorf("block = %d, expected > 5000000 for mainnet", block)
	}
	t.Logf("Latest block: %d", block)
}

// TestBeryxNetwork verifies that the network detection works correctly.
func TestBeryxNetwork(t *testing.T) {
	token := loadBeryxToken(t)

	client, err := NewBeryxClient(BeryxConfig{
		RPCURL:   "https://api.zondax.ch/fil/node/mainnet/rpc/v1",
		APIToken: token,
	})
	if err != nil {
		t.Fatalf("NewBeryxClient failed: %v", err)
	}
	defer client.Close()

	network := client.Network()
	if network != "mainnet" {
		t.Errorf("network = %q, want mainnet", network)
	}
	t.Logf("Network: %s", network)
}

// TestBeryxListProviders queries the on-chain SP Registry for active providers.
func TestBeryxListProviders(t *testing.T) {
	token := loadBeryxToken(t)

	client, err := NewBeryxClient(BeryxConfig{
		RPCURL:   "https://api.zondax.ch/fil/node/mainnet/rpc/v1",
		APIToken: token,
	})
	if err != nil {
		t.Fatalf("NewBeryxClient failed: %v", err)
	}
	defer client.Close()

	providers, err := client.ListProviders()
	if err != nil {
		t.Fatalf("ListProviders failed: %v", err)
	}
	t.Logf("Active providers: %d", len(providers))
	for _, p := range providers {
		t.Logf("  Provider #%d: %s (%s)", p.ID, p.Name, p.ServiceProvider.Hex())
	}
}

// TestBeryxUploadNotSupported verifies that upload returns a clear error.
func TestBeryxUploadNotSupported(t *testing.T) {
	token := loadBeryxToken(t)

	client, err := NewBeryxClient(BeryxConfig{
		RPCURL:   "https://api.zondax.ch/fil/node/mainnet/rpc/v1",
		APIToken: token,
	})
	if err != nil {
		t.Fatalf("NewBeryxClient failed: %v", err)
	}
	defer client.Close()

	_, err = client.Upload([]byte(`{"score":85}`))
	if err == nil {
		t.Error("expected error for upload via Beryx RPC")
	}
	t.Logf("Upload correctly rejected: %v", err)
}
