# Blockchain Anchoring

> **Status:** Planned (Phase 4)

The blockchain layer stores CID hashes on-chain and mints verifiable proof credentials (NFTs) for evaluated submissions. Uses Flow blockchain with Cadence smart contracts.

## Technology

| Component | Technology | Purpose |
|---|---|---|
| Chain | [Flow](https://developers.flow.com/) | Proof anchoring and credential minting |
| Contract language | [Cadence](https://cadence-lang.org/) | Smart contract logic |
| Integration | Go service in `backend/blockchain/` | Transaction submission and verification |
| Testnet | Flow Testnet | Development and testing |

## Smart Contract

```cadence
pub contract Proof {
    pub struct Credential {
        pub let cid: String
        pub let score: Int
        pub let timestamp: UFix64
    }

    pub fun anchorProof(cid: String, score: Int): Credential
    pub fun verifyProof(cid: String): Credential?
}
```

## How It Will Work

```mmd
AI Evaluation → IPFS Upload → CID → Flow Transaction → Proof NFT → Verification Link
```

1. Backend receives evaluation result with CID from storage layer
2. A Cadence transaction is submitted to the Proof contract with the CID and score
3. The contract stores the credential and mints a proof NFT
4. Transaction hash is returned to the user as verification
5. Anyone can verify by querying the contract with the CID

## Integration with Other Components

| Component | Interaction |
|---|---|
| **Impulse AI** | Provides the evaluation score anchored on-chain |
| **Filecoin (IPFS)** | Provides the CID of the stored evaluation report |
| **Lit Protocol** | Encrypts the report before IPFS storage; decryption keys tied to on-chain proof |
| **Flow** | Stores the CID hash and mints the verifiable credential |

## Environment Variables

| Variable | Description |
|---|---|
| `FLOW_ACCESS_NODE` | Flow access node URL (testnet: `https://rest-testnet.onflow.org`) |
| `FLOW_ACCOUNT_ADDRESS` | Flow account address for submitting transactions |
| `FLOW_PRIVATE_KEY` | Flow account private key |
| `CONTRACT_ADDRESS` | Deployed Proof contract address |

## Testing

```bash
# Contract tests (when implemented)
cd contracts && flow test

# Backend integration tests
cd backend && go test -v ./blockchain/
```

## Related Issues

- **#10:** Deploy proof credential smart contract
- **#11:** Backend integration for on-chain anchoring
