# Decentralized Storage

> **Status:** Implemented (Phase 3) — Lighthouse live, Beryx connected, CID management + cache

The storage layer persists AI evaluation reports on IPFS/Filecoin and returns content-addressed CIDs for blockchain anchoring.

## Technology

| Component | Technology | Purpose |
|---|---|---|
| Primary storage | [Lighthouse.storage](https://docs.lighthouse.storage/) | Stable Filecoin/IPFS pinning — upload + retrieve |
| Chain access | [Beryx (Zondax)](https://docs.zondax.ch/beryx) | Authenticated Filecoin RPC + data API |
| Storage SDK | [go-synapse](https://github.com/data-preservation-programs/go-synapse) | Native Filecoin PDP (fallback, when providers register) |
| Local storage | IPFS HTTP API | Development — upload/retrieve via local node |
| Language | Go | stdlib `net/http` + go-synapse + go-ethereum |

## Directory Structure

```md
backend/storage/
├── ipfs.go                      # Store interface, LighthouseClient, BeryxClient, FilecoinClient, IPFSClient, factory
├── cid.go                       # ComputeCID, ValidateCID, VerifyContent
├── cache.go                     # CachedStore (LRU cache wrapping any Store)
├── ipfs_test.go                 # 14 unit tests
├── cid_cache_test.go            # 14 unit tests (CID + cache)
├── beryx_integration_test.go    # 5 integration tests (live Beryx RPC)
└── lighthouse_integration_test.go # 2 integration tests (live Lighthouse upload/retrieve)
```

## Architecture

### Store Interface

```go
type Store interface {
    Upload(report []byte) (cid string, err error)
    Retrieve(cid string) (data []byte, err error)
}
```

### Backend Priority

```
NewStore(ipfsURL, lighthouseCfg, filecoinCfg)
  1. Lighthouse (if LIGHTHOUSE_API_KEY set)  — primary, stable
  2. go-synapse (if FILECOIN_PRIVATE_KEY set) — fallback, native PDP
  3. Local IPFS (default)                     — development
```

| Backend | Status | Auth | Operations |
|---|---|---|---|
| **LighthouseClient** | Live — roundtrip verified | Bearer API key | Upload + Retrieve |
| **BeryxClient** | Live — chain ID 314 | JWT Bearer | Read-only chain queries |
| **FilecoinClient** | SDK ready, 0 providers | ECDSA key | Upload + Retrieve (when available) |
| **IPFSClient** | Works locally | None | Upload + Retrieve |

### LighthouseClient (Primary)

Stable Filecoin/IPFS pinning with active storage providers:

```go
client := NewLighthouseClient(LighthouseConfig{APIKey: os.Getenv("LIGHTHOUSE_API_KEY")})
cid, _ := client.Upload(reportJSON)   // POST upload.lighthouse.storage/api/v0/add
data, _ := client.Retrieve(cid)       // GET gateway.lighthouse.storage/ipfs/{cid}
```

Verified live: upload 67 bytes, retrieve 67 bytes, content matches.

### BeryxClient (Chain Queries)

Authenticated Filecoin RPC via Zondax:

```go
client, _ := NewBeryxClient(BeryxConfig{
    RPCURL:   "https://api.zondax.ch/fil/node/mainnet/rpc/v1",
    APIToken: os.Getenv("BERYX_API_TOKEN"),
})
chainID := client.ChainID()           // 314 (mainnet)
providers, _ := client.ListProviders() // on-chain SP registry
```

### CID Management

```go
// Deterministic CID generation (SHA-256, CIDv1)
cid, _ := ComputeCID(reportJSON)

// Validate CID format before blockchain anchoring
normalized, _ := ValidateCID(cidStr)

// Verify content integrity after retrieval
match, _ := VerifyContent(cidStr, retrievedData)
```

### Cache Layer

```go
// Wrap any Store with LRU cache
cached := NewCachedStore(store, &CacheConfig{MaxSize: 1000, TTL: time.Hour})
cached.Upload(report)     // caches result
cached.Retrieve(cid)      // checks cache first (0 network calls on hit)
```

## Environment Variables

| Variable | Description | Example |
|---|---|---|
| `LIGHTHOUSE_API_KEY` | Lighthouse.storage API key | `7...` |
| `IPFS_API_URL` | Local IPFS node URL | `http://localhost:5001` |
| `BERYX_API_TOKEN` | Beryx JWT token | `eyJhbG...` |
| `FILECOIN_RPC_URL` | Filecoin RPC (Beryx or Glif) | `https://api.zondax.ch/fil/node/mainnet/rpc/v1` |
| `FILECOIN_DATA_URL` | Beryx data API | `https://api.zondax.ch/fil/data/v4/mainnet` |
| `FILECOIN_PROVIDER_URL` | go-synapse storage provider URL | — |
| `FILECOIN_PRIVATE_KEY` | ECDSA hex key for go-synapse | — |

## Testing

### Unit tests (no network needed)

```bash
cd backend && go test -race -v ./storage/
# 28 tests (14 storage + 14 CID/cache), ~1s
```

### Integration tests (require API keys)

```bash
# Lighthouse live roundtrip (requires LIGHTHOUSE_API_KEY in .env)
cd backend && go test -tags=integration -v -run TestLighthouse ./storage/

# Beryx live chain queries (requires BERYX_API_TOKEN)
export BERYX_API_TOKEN=your-jwt-token
cd backend && go test -tags=integration -v -run TestBeryx ./storage/
```

### E2E test script

```bash
./scripts/test_e2e_storage.sh
```

### Test Summary (35 total)

| Test | Type | Verifies |
|---|---|---|
| `TestIPFSUpload/Error/EmptyCID` | Unit | IPFS upload, errors, empty hash |
| `TestIPFSRetrieve/Error` | Unit | IPFS fetch, 404 handling |
| `TestLighthouseUpload/Error/Retrieve/Defaults` | Unit | Lighthouse auth, CID, gateway, URLs |
| `TestNewStore*` | Unit | Factory: IPFS, Lighthouse, priority |
| `TestFilecoinConfigValidation` | Unit | Invalid key rejected |
| `TestComputeCID*` | Unit | Deterministic, different content, empty |
| `TestValidateCID*` | Unit | Valid, empty, invalid |
| `TestVerifyContent*` | Unit | Match, mismatch |
| `TestCachedStore*` | Unit | Upload, hit, miss, eviction, TTL, stats |
| `TestBeryx*` | Integration | Connection, block, network, providers, upload |
| `TestLighthouse*Live/Roundtrip` | Integration | Live upload + retrieve, content match |

## Contract Addresses (auto-resolved by go-synapse)

| Contract | Mainnet (314) | Calibration (314159) |
|---|---|---|
| PDPVerifier | `0xBADd0B92C1c71d02E7d520f64c0876538fa2557F` | `0x85e366Cf9DD2c0aE37E963d9556F5f4718d6417C` |
| SP Registry | `0xf55dDbf63F1b55c3F1D4FA7e339a68AB7b64A5eB` | `0x839e5c9988e4e9977d40708d0094103c0839Ac9D` |

## Current Status

- **Lighthouse:** Primary backend — live upload/retrieve roundtrip verified
- **Beryx RPC:** Connected to Filecoin Mainnet (chain ID 314, block 5.8M+)
- **go-synapse:** SDK integrated, 0 Synapse providers registered — Lighthouse handles storage
- **CID management:** Deterministic generation, validation, content verification
- **Cache:** LRU with configurable size (1000) and TTL (1 hour)
- **Factory priority:** Lighthouse > go-synapse > local IPFS
