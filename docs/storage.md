# Decentralized Storage

> **Status:** Implemented (Phase 3)

The storage layer persists AI evaluation reports on IPFS/Filecoin and returns content-addressed CIDs for blockchain anchoring. Supports both a local IPFS node (development) and Filecoin via web3.storage (production).

## Technology

- **Storage:** IPFS (content-addressed) + Filecoin (long-term persistence)
- **Backends:** Local IPFS node (HTTP API) or [web3.storage](https://web3.storage/) (IPFS pinning + Filecoin deals)
- **Language:** Go (stdlib `net/http`, no external dependencies)
- **Package:** `backend/storage/`

## Directory Structure

```md
backend/storage/
├── ipfs.go          # Store interface, IPFSClient, FilecoinClient, factory
└── ipfs_test.go     # 15 tests with HTTP test server mocks
```

## Architecture

### Store Interface

```go
type Store interface {
    Upload(report []byte) (cid string, err error)
    Retrieve(cid string) (data []byte, err error)
}
```

Both backends implement this interface, making them swappable via the factory function.

### Backend Selection

```mmd
NewStore(ipfsURL, filecoinAPIURL, filecoinGatewayURL, filecoinToken)
  ├── filecoinToken set? → FilecoinClient (web3.storage)
  └── otherwise          → IPFSClient (default: localhost:5001)
```

### Upload Flow

```mmd
Report JSON → Multipart Form → IPFS/Filecoin API → CID returned
```

1. Report bytes are wrapped in a multipart form upload (field: "file", name: "report.json")
2. Sent to the storage backend via HTTP POST
3. Backend returns a CID (content hash)
4. CID is passed to the blockchain layer for on-chain anchoring

**Filecoin dual-storage:** When using web3.storage, data is pinned to IPFS immediately for fast retrieval, and a Filecoin storage deal is created asynchronously for long-term persistence.

### Retrieve Flow

```mmd
CID → IPFS cat / w3s.link gateway → Report JSON returned
```

## API Reference (Internal)

### IPFSClient

| Method | Endpoint | Description |
|---|---|---|
| Upload | `POST /api/v0/add` | Multipart upload, returns `{"Hash": "Qm..."}` |
| Retrieve | `POST /api/v0/cat?arg={cid}` | Returns raw report bytes |

### FilecoinClient (web3.storage)

| Method | Endpoint | Description |
|---|---|---|
| Upload | `POST /upload` | Multipart upload with Bearer auth, returns `{"cid": "bafy..."}` |
| Retrieve | `GET {gateway}/ipfs/{cid}` | Gateway fetch (default: w3s.link) |

## Environment Variables

| Variable | Description | Default |
|---|---|---|
| `IPFS_API_URL` | Local IPFS node HTTP API URL | `http://localhost:5001` |
| `FILECOIN_API_URL` | web3.storage API base URL | `https://api.web3.storage` |
| `FILECOIN_GATEWAY_URL` | IPFS/Filecoin gateway for retrieval | `https://w3s.link` |
| `FILECOIN_TOKEN` | web3.storage API token (if set, Filecoin is used instead of local IPFS) | — |

## Error Handling

| Error | Cause | Handling |
|---|---|---|
| Upload returns non-200 | IPFS node down, auth failure | Error with status code and body |
| Empty CID in response | Malformed response | Error: "empty CID in response" |
| Retrieve returns non-200 | CID not found, network error | Error with status code and body |
| HTTP client timeout | Network issues (30s IPFS, 60s Filecoin) | Error wrapping the timeout |

## Running

```bash
# With local IPFS node (development)
export IPFS_API_URL=http://localhost:5001
cd backend && go run main.go

# With Filecoin via web3.storage (production)
export FILECOIN_TOKEN=your-web3storage-token
cd backend && go run main.go
```

## Testing

```bash
cd backend && go test -race -v ./storage/
# 15 tests, ~1s
```

All tests use `httptest.NewServer` to mock both APIs — no real IPFS node or web3.storage account needed.

### Test Summary

| Test | Verifies |
|---|---|
| `TestIPFSUpload` | Multipart upload, CID extraction from Hash field |
| `TestIPFSUploadError` | Error handling for 500 response |
| `TestIPFSUploadEmptyCID` | Error handling for empty Hash |
| `TestIPFSRetrieve` | Fetch by CID, correct query parameter |
| `TestIPFSRetrieveError` | Error handling for 404 response |
| `TestFilecoinUpload` | Bearer auth header, CID from cid field, multipart body |
| `TestFilecoinUploadError` | Error handling for 401 response |
| `TestFilecoinUploadEmptyCID` | Error handling for empty cid |
| `TestFilecoinRetrieve` | Gateway fetch with correct path |
| `TestFilecoinRetrieveError` | Error handling for 404 response |
| `TestNewStoreIPFS` | Factory returns IPFSClient when no Filecoin token |
| `TestNewStoreFilecoin` | Factory returns FilecoinClient when token is set |
| `TestNewStoreDefaultIPFS` | Factory uses default localhost:5001 |
| `TestNewFilecoinClientDefaults` | Default API and gateway URLs |
