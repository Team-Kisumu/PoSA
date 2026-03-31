# Blockchain Anchoring

> **Status:** Deployed on Flow Testnet (Phase 4)

The blockchain layer stores CID hashes on-chain and mints verifiable proof credentials. The ProofOfSkill contract is live on Flow Testnet.

## Live Deployment

| Field | Value |
|---|---|
| Contract | `ProofOfSkill` |
| Network | Flow Testnet (chain: `flow-testnet`) |
| Account | [`0x***`](https://testnet.flowscan.io/account/0xf8a2fcf3389475a1) |
| Key | ECDSA_P256 / SHA3_256 (index 0, weight 1000) |
| Deploy tx | [`ae...`](https://testnet.flowscan.io/tx/aea0e47f0511db4b4e0b9eb13ad240738aa1d620e7055304ea47ca47d0712d16) |
| First mint tx | [`01d***...`](https://testnet.flowscan.io/tx/01dcef77893adb7d4f124a1ecc48d77834f24fbd5e1a91fbb055ffc9e084236f) |

## Technology

| Component | Technology | Purpose |
|---|---|---|
| Chain | [Flow](https://developers.flow.com/) | Proof anchoring and credential minting |
| Contract language | [Cadence](https://cadence-lang.org/) | Smart contract logic |
| CLI | [Flow CLI v2.15.3](https://github.com/onflow/flow-cli) | Deploy, test, interact |
| Go integration | `backend/blockchain/` | Transaction submission and verification |
| REST API | `https://rest-testnet.onflow.org/v1` | Public, no auth needed |

## Directory Structure

```md
contracts/
├── ProofOfSkill.cdc                    # Main contract
├── ProofOfSkill_test.cdc               # 10 Cadence tests
├── transactions/
│   └── mint_credential.cdc             # Mint transaction (admin only)
└── scripts/
    ├── verify_credential.cdc           # Read-only CID lookup
    └── get_total_minted.cdc            # Counter query

backend/blockchain/
├── blockchain.go                       # FlowClient (AnchorProof, VerifyProof, HasProof)
└── blockchain_test.go                  # 9 Go tests with HTTP mocks

flow.json                               # Flow project config
```

## Smart Contract — ProofOfSkill

```cadence
access(all) contract ProofOfSkill {

    access(all) struct Credential {
        access(all) let id: UInt64
        access(all) let cid: String
        access(all) let score: UInt8        // 0-100
        access(all) let submitter: Address
        access(all) let timestamp: UFix64
    }

    access(all) resource Admin {
        access(all) fun mintCredential(cid: String, score: UInt8, submitter: Address): Credential
    }

    access(all) fun getCredential(cid: String): Credential?
    access(all) fun getCredentialByID(id: UInt64): Credential?
    access(all) fun hasCredential(cid: String): Bool
    access(all) fun getTotalMinted(): UInt64
}
```

### Security

- **Admin-restricted minting:** Only the contract deployer holds the Admin resource
- **Duplicate prevention:** Rejects minting if CID already exists on-chain
- **Score validation:** UInt8 enforces 0-255 range; precondition enforces 0-100
- **Events:** `CredentialMinted` emitted on every mint for monitoring

## Integration with Pipeline

```mmd
User Upload → Go Backend → Impulse AI (score) → Lighthouse (CID) → Flow (anchor) → Verification
```

| Step | Component | Output |
|---|---|---|
| 1. Submit code | Go Backend | Validated file |
| 2. Analyze | Impulse AI + local rules | Score 0-100, issues, suggestions |
| 3. Store report | Lighthouse.storage | CID (e.g. `QmXZh24F...`) |
| 4. Anchor proof | Flow ProofOfSkill | Transaction hash, credential ID |
| 5. Verify | Anyone queries Flow | Credential with CID, score, timestamp |

## Environment Variables

| Variable | Description | Value |
|---|---|---|
| `FLOW_ACCESS_NODE` | Flow REST API (public, no auth) | `https://rest-testnet.onflow.org` |
| `FLOW_ACCOUNT_ADDRESS` | Deployer account | `0xf8a2fcf3389475a1` |
| `FLOW_PRIVATE_KEY` | ECDSA_P256 hex key for signing | (in .env, gitignored) |
| `CONTRACT_ADDRESS` | ProofOfSkill contract address | `0xf8a2fcf3389475a1` |

### Flow Access API — No Auth Required

The Flow REST API is completely public:

- **Testnet:** `https://rest-testnet.onflow.org/v1`
- **Mainnet:** `https://rest-mainnet.onflow.org/v1`

No API key, no Bearer token, no authentication of any kind.

### Account Creation

Flow testnet accounts are created via the faucet API:

```bash
# Create account (returns new address)
curl -X POST https://faucet.flow.com/api/account \
  -H "Content-Type: application/json" \
  -d '{"publicKey":"<hex>","signatureAlgorithm":"ECDSA_P256","hashAlgorithm":"SHA3_256"}'

# Fund account with testnet FLOW
curl -X POST https://faucet.flow.com/api/fund \
  -H "Content-Type: application/json" \
  -d '{"address":"0x...","token":"FLOW"}'
```

## Deployment

### Prerequisites

```bash
# Install Flow CLI
curl -fsSL "https://github.com/onflow/flow-cli/releases/download/v2.15.3/flow-cli-v2.15.3-linux-amd64.tar.gz" \
  | tar -xz -C ~/bin/
```

### Deploy Contract

```bash
export FLOW_PRIVATE_KEY=<your-hex-key>
flow project deploy --network testnet
```

### Mint a Credential

```bash
flow transactions send contracts/transactions/mint_credential.cdc \
  "QmYourCIDHere" 85 0xf8a2fcf3389475a1 \
  --network testnet --signer testnet-account
```

### Verify a Credential

```bash
flow scripts execute contracts/scripts/verify_credential.cdc \
  "QmYourCIDHere" --network testnet
```

### Query Total Minted

```bash
flow scripts execute contracts/scripts/get_total_minted.cdc --network testnet
```

## Testing

### Cadence Tests (require Flow CLI)

```bash
flow test contracts/ProofOfSkill_test.cdc
```

10 test cases: initial state, mint+verify, hasCredential, getByID, duplicate rejection, score validation, empty CID, sequential IDs.

### Go Tests

```bash
cd backend && go test -race -v ./blockchain/
```

9 tests with HTTP mocks: input validation, successful anchor, server error, verify found/not found, HasProof, client creation.

### Live Testnet Verification

```bash
# Verified on-chain:
# getTotalMinted() = 1
# getCredential("QmTestCID...") = Credential{id:1, score:85, submitter:0xf8a2...}
# CredentialMinted event emitted at block 313,938,615
```

## Block Explorer

View all transactions and contract state:

- Account: <https://testnet.flowscan.io/account/0xf8a2fcf3389475a1>
- Deploy tx: <https://testnet.flowscan.io/tx/aea0e47f0511db4b4e0b9eb13ad240738aa1d620e7055304ea47ca47d0712d16>
- Mint tx: <https://testnet.flowscan.io/tx/01dcef77893adb7d4f124a1ecc48d77834f24fbd5e1a91fbb055ffc9e084236f>
