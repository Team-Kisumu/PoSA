# Decentralized Storage

> **Status:** Implemented (Phase 3)

The storage layer persists AI evaluation reports on IPFS/Filecoin and returns content-addressed CIDs for blockchain anchoring.

## Technology

| Component | Technology | Purpose |
|---|---|---|
| Local storage | IPFS HTTP API | Development — upload/retrieve via local node |
| Chain access | [Beryx (Zondax)](https://docs.zondax.ch/beryx) | Authenticated Filecoin RPC + data API |
| Storage SDK | [go-synapse](https://github.com/data-preservation-programs/go-synapse) | Native Filecoin storage with PDP (when providers available) |
| Language | Go | stdlib `net/http` + go-synapse + go-ethereum |

## Directory Structure

```md
backend/storage/
├── ipfs.go                      # Store interface, BeryxClient, FilecoinClient, IPFSClient, factory
├── ipfs_test.go                 # 8 unit tests (IPFS mocks + factory + config validation)
└── beryx_integration_test.go    # 5 integration tests (live Beryx RPC, build tag: integration)
```

## Architecture

### Store Interface

```go
type Store interface {
    Upload(report []byte) (cid string, err error)
    Retrieve(cid string) (data []byte, err error)
}
```

### Three Backends

```mmd
                    ┌─────────────────────────────────────┐
                    │           Store Interface            │
                    │       Upload() / Retrieve()          │
                    └──────┬──────────┬──────────┬────────┘
                           │          │          │
                    ┌──────▼──┐ ┌─────▼────┐ ┌──▼────────┐
                    │  Beryx  │ │ Filecoin │ │   IPFS    │
                    │ Client  │ │  Client  │ │  Client   │
                    │ (chain) │ │(storage) │ │  (local)  │
                    └─────────┘ └──────────┘ └───────────┘
```

| Backend | Use Case | Auth | Data Operations |
|---|---|---|---|
| **BeryxClient** | Chain queries, provider discovery | JWT Bearer token | Read-only (ChainID, BlockNumber, ListProviders) |
| **FilecoinClient** | Production storage (when providers register) | ECDSA private key | Upload (UploadBytes) + Retrieve (Download) |
| **IPFSClient** | Local development | None | Upload (/api/v0/add) + Retrieve (/api/v0/cat) |

### BeryxClient (Zondax)

Authenticated access to Filecoin via Beryx RPC proxy:

```go
client, err := NewBeryxClient(BeryxConfig{
    RPCURL:   "https://api.zondax.ch/fil/node/mainnet/rpc/v1",
    APIToken: os.Getenv("BERYX_API_TOKEN"),
})
// Chain queries
chainID := client.ChainID()       // 314 (mainnet)
block, _ := client.BlockNumber()   // latest height
providers, _ := client.ListProviders() // on-chain SP registry
```

Beryx endpoints:
- **RPC proxy:** `https://api.zondax.ch/fil/node/mainnet/rpc/v1`
- **Data API:** `https://api.zondax.ch/fil/data/v4/mainnet`
- **Calibration RPC:** `https://api.zondax.ch/fil/node/calibration/rpc/v1`

### FilecoinClient (go-synapse)

Native Filecoin storage via the Synapse protocol:

```go
client, err := NewFilecoinClient(FilecoinConfig{
    PrivateKeyHex: os.Getenv("FILECOIN_PRIVATE_KEY"),
    RPCURL:        os.Getenv("FILECOIN_RPC_URL"),
    ProviderURL:   os.Getenv("FILECOIN_PROVIDER_URL"),
})
cid, _ := client.Upload(reportJSON)
data, _ := client.Retrieve(cid)
```

### IPFSClient (Local)

For development without Filecoin:

```go
client := NewIPFSClient("http://localhost:5001")
cid, _ := client.Upload(reportJSON)   // POST /api/v0/add
data, _ := client.Retrieve(cid)       // POST /api/v0/cat?arg={cid}
```

## Environment Variables

| Variable | Description | Example |
|---|---|---|
| `IPFS_API_URL` | Local IPFS node URL | `http://localhost:5001` |
| `BERYX_API_TOKEN` | Beryx JWT token for authenticated RPC | `eyJhbG...` |
| `FILECOIN_RPC_URL` | Filecoin RPC (Beryx or Glif) | `https://api.zondax.ch/fil/node/mainnet/rpc/v1` |
| `FILECOIN_DATA_URL` | Beryx data API | `https://api.zondax.ch/fil/data/v4/mainnet` |
| `FILECOIN_PROVIDER_URL` | go-synapse storage provider URL | — (none registered yet) |
| `FILECOIN_PRIVATE_KEY` | ECDSA hex key for go-synapse transactions | — |

### RPC Endpoints

| Provider | Network | URL | Auth |
|---|---|---|---|
| Beryx (Zondax) | Mainnet | `https://api.zondax.ch/fil/node/mainnet/rpc/v1` | Bearer JWT |
| Beryx (Zondax) | Calibration | `https://api.zondax.ch/fil/node/calibration/rpc/v1` | Bearer JWT |
| Glif (public) | Mainnet | `https://api.node.glif.io/rpc/v1` | None |
| Glif (public) | Calibration | `https://api.calibration.node.glif.io/rpc/v1` | None |

## Running

```bash
# Local development (IPFS)
export IPFS_API_URL=http://localhost:5001
cd backend && go run main.go

# With Beryx chain access
export BERYX_API_TOKEN=your-jwt-token
export FILECOIN_RPC_URL=https://api.zondax.ch/fil/node/mainnet/rpc/v1
cd backend && go run main.go
```

## Testing

### Unit tests (no network needed)

```bash
cd backend && go test -race -v ./storage/
# 8 tests, ~1s
```

### Integration tests (requires BERYX_API_TOKEN)

```bash
export BERYX_API_TOKEN=your-jwt-token
cd backend && go test -tags=integration -v ./storage/
# 13 tests (8 unit + 5 integration), ~9s
```

### E2E test script

```bash
./scripts/test_e2e_storage.sh
```

### Test Summary

| Test | Type | Verifies |
|---|---|---|
| `TestIPFSUpload` | Unit | Multipart upload, CID extraction |
| `TestIPFSUploadError` | Unit | Error handling for 500 |
| `TestIPFSUploadEmptyCID` | Unit | Empty Hash rejection |
| `TestIPFSRetrieve` | Unit | Fetch by CID |
| `TestIPFSRetrieveError` | Unit | Error handling for 404 |
| `TestNewStoreIPFS` | Unit | Factory returns IPFSClient |
| `TestNewStoreDefaultIPFS` | Unit | Default localhost:5001 |
| `TestFilecoinConfigValidation` | Unit | Invalid private key rejected |
| `TestBeryxConnection` | Integration | Chain ID = 314 (mainnet) |
| `TestBeryxBlockNumber` | Integration | Block > 5,000,000 |
| `TestBeryxNetwork` | Integration | Network = mainnet |
| `TestBeryxListProviders` | Integration | SP Registry query |
| `TestBeryxUploadNotSupported` | Integration | Clear error for upload via RPC |

## Contract Addresses (auto-resolved by go-synapse)

| Contract | Mainnet (314) | Calibration (314159) |
|---|---|---|
| PDPVerifier | `0xBADd0B92C1c71d02E7d520f64c0876538fa2557F` | `0x85e366Cf9DD2c0aE37E963d9556F5f4718d6417C` |
| SP Registry | `0xf55dDbf63F1b55c3F1D4FA7e339a68AB7b64A5eB` | `0x839e5c9988e4e9977d40708d0094103c0839Ac9D` |
| Payments | `0x23b1e018F08BB982348b15a86ee926eEBf7F4DAa` | `0x09a0fDc2723fAd1A7b8e3e00eE5DF73841df55a0` |
| FWSS | `0x8408502033C418E1bbC97cE9ac48E5528F371A9f` | `0x02925630df557F957f70E112bA06e50965417CA0` |
| SessionKeyRegistry | `0x74FD50525A958aF5d484601E252271f9625231aB` | `0x518411c2062E119Aaf7A8B12A2eDf9a939347655` |

## Current Status

- **IPFS local:** Fully functional for development
- **Beryx RPC:** Connected to Filecoin Mainnet (chain ID 314, block 5.8M+)
- **go-synapse storage:** SDK integrated, awaiting storage provider registration on Synapse protocol
- **Provider discovery:** `BeryxClient.ListProviders()` queries on-chain SP Registry (currently 0 active providers)
