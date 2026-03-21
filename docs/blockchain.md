# Blockchain Anchoring

> **Status:** 📋 Planned (Phase 4)

The blockchain layer stores CID hashes on-chain and mints verifiable proof credentials (NFTs) for evaluated submissions.

## Technology

- **Chain:** Flow or NEAR
- **Contract language:** Cadence (Flow) / Rust (NEAR)
- **Integration:** Go service in `backend/blockchain/`

## Smart Contract

```cadence
pub contract Proof {
    pub struct Credential {
        pub let cid: String
        pub let score: Int
    }
}
```

## How It Will Work

1. Backend receives a CID from the storage layer
2. A transaction is submitted to the smart contract with the CID and score
3. The contract stores the credential and mints a proof NFT
4. Transaction hash is returned to the user as verification

## Planned API (Internal Service)

```go
// blockchain/blockchain.go
func AnchorProof(cid string, score int) (txHash string, err error)
func VerifyProof(cid string) (credential Credential, err error)
```

## Environment Variables

| Variable | Description |
|---|---|
| `BLOCKCHAIN_RPC` | Flow/NEAR RPC endpoint |
| `CONTRACT_ADDRESS` | Deployed smart contract address |

## Testing

```bash
# Contract tests
cd contracts && flow test  # (Flow)
# or
cd contracts && cargo test  # (NEAR)

# Integration tests
cd backend && go test -v ./blockchain/...
```

## Related Issues

- **#10:** Deploy proof credential smart contract
- **#11:** Backend integration for on-chain anchoring
