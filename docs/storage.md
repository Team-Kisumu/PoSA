# Decentralized Storage

> **Status:** Implemented (Phase 3)

The storage layer persists AI evaluation reports on IPFS/Filecoin and returns content-addressed CIDs for blockchain anchoring. Uses [go-synapse](https://github.com/data-preservation-programs/go-synapse) for native Filecoin integration with Proof of Data Possession (PDP).

## Technology

- **Storage:** IPFS (content-addressed) + Filecoin (long-term persistence with PDP)
- **SDK:** [go-synapse](https://github.com/data-preservation-programs/go-synapse) — Go SDK for Filecoin Synapse protocol
- **Backends:** Local IPFS node (development) or Filecoin via go-synapse (production)
- **Language:** Go
- **Package:** `backend/storage/`

## Directory Structure

```md
backend/storage/
├── ipfs.go          # Store interface, FilecoinClient (go-synapse), IPFSClient, factory
└── ipfs_test.go     # 8 tests (IPFS mocks + factory + config validation)
```

## Architecture

### Store Interface

```go
type Store interface {
    Upload(report []byte) (cid string, err error)
    Retrieve(cid string) (data []byte, err error)
}
```

### Backend Selection

```mmd
NewStore(ipfsURL, filecoinCfg)
  ├── filecoinCfg.PrivateKeyHex set? → FilecoinClient (go-synapse)
  └── otherwise                      → IPFSClient (default: localhost:5001)
```

### FilecoinClient (go-synapse)

The FilecoinClient uses go-synapse to interact with Filecoin storage providers:

- **Upload:** `storage.Manager.UploadBytes()` — sends data to the storage provider, returns a piece CID
- **Retrieve:** `storage.Manager.Download()` — fetches data by piece CID from the provider
- **PDP:** Proof of Data Possession — on-chain verification that the provider still holds the data
- **Network:** Supports Filecoin Mainnet (chain ID 314) and Calibration testnet (chain ID 314159)

```go
client, err := synapse.New(ctx, synapse.Options{
    PrivateKey:  privateKey,
    RPCURL:      "https://api.calibration.node.glif.io/rpc/v1",
    ProviderURL: "https://provider.example.com",
})
storage, _ := client.Storage()
result, _ := storage.UploadBytes(ctx, reportJSON, nil)
// result.PieceCID is the content identifier
```

### IPFSClient (HTTP API)

For local development without Filecoin:

| Method | Endpoint | Description |
|---|---|---|
| Upload | `POST /api/v0/add` | Multipart upload, returns `{"Hash": "Qm..."}` |
| Retrieve | `POST /api/v0/cat?arg={cid}` | Returns raw report bytes |

## Environment Variables

| Variable | Description | Default |
|---|---|---|
| `IPFS_API_URL` | Local IPFS node HTTP API URL | `http://localhost:5001` |
| `FILECOIN_RPC_URL` | Filecoin RPC endpoint | `https://api.calibration.node.glif.io/rpc/v1` |
| `FILECOIN_PROVIDER_URL` | Storage provider API endpoint | — |
| `FILECOIN_PRIVATE_KEY` | Hex-encoded ECDSA private key for transactions | — |

## Running

```bash
# Local development (IPFS node)
export IPFS_API_URL=http://localhost:5001
cd backend && go run main.go

# Production (Filecoin via go-synapse)
export FILECOIN_RPC_URL=https://api.calibration.node.glif.io/rpc/v1
export FILECOIN_PROVIDER_URL=https://your-provider.example.com
export FILECOIN_PRIVATE_KEY=your-hex-private-key
cd backend && go run main.go
```

## Testing

```bash
cd backend && go test -race -v ./storage/
# 8 tests, ~1s
```

### Test Summary

| Test | Verifies |
|---|---|
| `TestIPFSUpload` | Multipart upload, CID extraction from Hash field |
| `TestIPFSUploadError` | Error handling for 500 response |
| `TestIPFSUploadEmptyCID` | Error handling for empty Hash |
| `TestIPFSRetrieve` | Fetch by CID, correct query parameter |
| `TestIPFSRetrieveError` | Error handling for 404 response |
| `TestNewStoreIPFS` | Factory returns IPFSClient when no Filecoin config |
| `TestNewStoreDefaultIPFS` | Factory uses default localhost:5001 |
| `TestFilecoinConfigValidation` | Invalid private key rejected |

### Integration Testing

Full Filecoin integration tests require a Calibration testnet account:

```bash
export FILECOIN_RPC_URL=https://api.calibration.node.glif.io/rpc/v1
export FILECOIN_PROVIDER_URL=https://your-provider.example.com
export FILECOIN_PRIVATE_KEY=your-test-private-key
cd backend && go test -tags=integration -v ./storage/
```

## Contract Addresses (go-synapse)

| Network | PDPVerifier Contract |
|---|---|
| Filecoin Mainnet (314) | `0xBADd0B92C1c71d02E7d520f64c0876538fa2557F` |
| Filecoin Calibration (314159) | `0x85e366Cf9DD2c0aE37E963d9556F5f4718d6417C` |
