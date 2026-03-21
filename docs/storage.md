# Decentralized Storage

> **Status:** 📋 Planned (Phase 3)

The storage layer persists AI evaluation reports on IPFS/Filecoin and returns content-addressed CIDs for blockchain anchoring.

## Technology

- **Storage:** IPFS / Filecoin
- **Gateway:** Local IPFS node or [Pinata](https://www.pinata.cloud/) API
- **Integration:** Go service in `backend/services/`

## How It Will Work

1. Backend receives an evaluation report from the AI engine
2. Report is serialized to JSON and uploaded to IPFS
3. IPFS returns a CID (content identifier) — a cryptographic hash of the report
4. CID is passed to the blockchain layer for on-chain anchoring
5. Reports are retrievable by anyone with the CID

## Planned API (Internal Service)

```go
// services/storage.go
func UploadReport(report []byte) (cid string, err error)
func FetchReport(cid string) (report []byte, err error)
```

## Environment Variables

| Variable | Description |
|---|---|
| `IPFS_API_URL` | IPFS node or Pinata gateway URL |

## Testing

```bash
# Integration tests will verify:
# - Upload returns a valid CID
# - Fetch by CID returns the original report
# - Error handling for upload failures
# - CID validation before blockchain anchoring

cd backend && go test -v ./services/...
```

## Related Issues

- **#8:** IPFS/Filecoin integration for report persistence
- **#9:** CID generation, retrieval, and caching
