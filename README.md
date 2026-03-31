# Proof-of-Skill AI (PoSA)

### AI-Verified Work. Blockchain-Trusted Proof

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

---

## Table of Contents

- [Overview](#overview)
- [Problem Statement](#problem-statement)
- [Solution](#solution)
- [Architecture](#architecture)
- [Orchestration Flow](#orchestration-flow)
- [Project Structure](#project-structure)
- [Implementation Strategy](#implementation-strategy)
- [Getting Started](#getting-started)
- [How to Use](#how-to-use)
- [Example Runs](#example-runs)
- [Security](#security)
- [CI/CD & Automations](#cicd--automations)
- [Use Cases](#use-cases)
- [Future Enhancements](#future-enhancements)
- [Contributing](#contributing)
- [License](#license)

---

## Overview

Proof-of-Skill AI (PoSA) is a decentralized system that uses **Artificial Intelligence** to evaluate user-submitted work (code, writing, tasks) and **Blockchain** to issue tamper-proof, verifiable credentials.

It addresses the growing problem of trust in digital work, freelancing, and online credentials — bridging the gap between skill and trust in the digital economy.

---

## Problem Statement

- Fake portfolios and unverifiable claims plague hiring pipelines
- No standardized, trusted proof of real skills exists
- Assessing remote talent remains difficult and subjective
- Security audits lack transparency and immutability

---

## Solution

PoSA introduces a pipeline where:

1. **AI evaluates** submitted work (code quality, security, logic)
2. **Reports are stored** on decentralized storage (IPFS/Filecoin)
3. **A cryptographic hash (CID)** is recorded on-chain
4. **A verifiable credential** (NFT or proof record) is issued to the user

---

## Architecture

```mermaid
flowchart TD
    A[User Upload] --> B[Backend - Go API]
    B --> C[AI Engine]
    C --> D[Evaluation Report]
    D --> E[IPFS / Filecoin Storage]
    E --> F[CID Hash]
    F --> G[Blockchain - Flow/NEAR]
    G --> H[Proof Credential NFT]
    H --> I[User Verification Link]
```

### Components

| Component | Technology | Responsibility | Status |
|---|---|---|---|
| **Frontend** | React / Next.js | Upload interface, displays AI evaluation results | Planned |
| **Backend** | Go 1.25+ | Handles submissions, orchestrates AI → Storage → Blockchain | Implemented |
| **AI Engine** | Python + [Impulse AI](https://docs.impulselabs.ai/) | Pattern-based analysis (25 langs) + AI-powered evaluation | Implemented |
| **Storage** | [Lighthouse.storage](https://docs.lighthouse.storage/) + [Beryx](https://docs.zondax.ch/beryx) | Stores evaluation reports on IPFS/Filecoin, returns CID | Live |
| **Blockchain** | [Flow](https://developers.flow.com/) (Cadence) | Stores CID hash, mints proof credential | [Deployed on Testnet](https://testnet.flowscan.io/account/0xf8a2fcf3389475a1) |

---

## Orchestration Flow

```mermaid
sequenceDiagram
    participant U as User
    participant B as Backend
    participant AI as AI Engine
    participant IPFS as Storage
    participant BC as Blockchain

    U->>B: Upload Work
    B->>AI: Send Data
    AI-->>B: Evaluation Result
    B->>IPFS: Store Report
    IPFS-->>B: CID
    B->>BC: Store Hash + Mint Proof
    BC-->>U: Transaction + Credential
```

**Simplified workflow:**

```mermaid
graph LR
    Upload --> Analyze --> Store --> Anchor --> Verify
```

---

## Project Structure

```txt
posa/
├── backend/
│   ├── main.go              # API entrypoint (:8080)
│   ├── handlers/            # HTTP route handlers + validation
│   ├── middleware/           # Security headers, request ID, recovery
│   ├── validation/           # Input validation (filename, MIME, content, CID)
│   ├── storage/             # Lighthouse, Beryx, go-synapse, IPFS, CID, cache
│   ├── blockchain/          # Flow integration (AnchorProof, VerifyProof)
│   └── services/            # Business logic orchestration
├── ai/
│   ├── evaluator.py         # FastAPI server (:8000)
│   ├── analyzer.py          # Routes to code or writing evaluator
│   ├── writing.py           # Writing quality evaluator
│   ├── impulse_client.py    # Impulse AI SSE streaming client
│   ├── rules/               # 25 language rule sets + cross-language rules
│   └── tests/               # 150 pytest tests
├── contracts/
│   ├── ProofOfSkill.cdc     # Flow Cadence smart contract (deployed)
│   ├── transactions/        # Mint credential transaction
│   ├── scripts/             # Verify credential, get total minted
│   └── ProofOfSkill_test.cdc # 10 Cadence tests
├── docs/                    # Component documentation
├── scripts/                 # Test scripts, git hooks, setup
├── .env.example             # Environment config template
├── flow.json                # Flow project config
└── README.md
```

---

## Implementation Strategy

The project is built in **six incremental phases**:

### Phase 1 — Backend API (Go)

- REST API to accept file uploads and repo links
- Input validation and sanitization layer

### Phase 2 — AI Engine (Python)

- Code quality analysis (bugs, security vulnerabilities, logic correctness)
- Writing evaluation (quality, originality)
- Scoring system (0–100) with detailed explanations

### Phase 3 — Decentralized Storage

- IPFS/Filecoin integration for report persistence
- CID generation and retrieval

### Phase 4 — Blockchain Anchoring

- Smart contract deployment (Flow or NEAR)
- CID hash written on-chain, credential minted

```cadence
pub contract Proof {
    pub struct Credential {
        pub let cid: String
        pub let score: Int
    }
}
```

### Phase 5 — Frontend

- Upload interface (file, paste, GitHub link)
- Real-time AI feedback display
- Proof minting and verification UI

### Phase 6 — Integration & Hardening

- End-to-end pipeline testing
- Security audits, input fuzzing, contract verification

---

## Getting Started

### Prerequisites

- [Go](https://go.dev/) (1.25+)
- [Python](https://www.python.org/) (3.10+)
- [Node.js](https://nodejs.org/) (18+) — for frontend (planned)
- [Flow CLI](https://github.com/onflow/flow-cli) (v2.15+) — for contract deployment

### Setup

```bash
# Clone the repository
git clone https://github.com/Team-Kisumu/PoSA.git
cd PoSA

# Configure environment
cp .env.example .env
# Edit .env with your API keys (see Environment Variables below)

# Install git hooks
./scripts/setup-hooks.sh

# Backend (terminal 1)
cd backend
go run main.go
# -> PoSA backend listening on :8080

# AI Engine (terminal 2)
cd ai
python3 -m venv .venv
source .venv/bin/activate
pip install -r requirements.txt
PYTHONPATH=.. python evaluator.py
# -> Uvicorn running on http://0.0.0.0:8000

# Deploy smart contract (one-time)
export FLOW_PRIVATE_KEY=<your-key>
flow project deploy --network testnet
```

### Running Tests

```bash
# All backend tests
cd backend && go test -race ./...

# All AI tests
source ai/.venv/bin/activate && PYTHONPATH=. pytest ai/tests/ -q

# AI end-to-end (requires AI_API_KEY in .env)
PYTHONPATH=. python scripts/test_e2e_ai.py

# Storage integration (requires LIGHTHOUSE_API_KEY)
cd backend && go test -tags=integration -v ./storage/

# Blockchain integration (requires FLOW_PRIVATE_KEY)
export FLOW_PRIVATE_KEY=<key>
flow scripts execute contracts/scripts/get_total_minted.cdc --network testnet
```

### Environment Variables

| Variable | Description |
|---|---|
| `AI_API_KEY` | [Impulse AI](https://docs.impulselabs.ai/) API key (x-api-key auth) |
| `LIGHTHOUSE_API_KEY` | [Lighthouse.storage](https://docs.lighthouse.storage/) API key (primary storage) |
| `IPFS_API_URL` | Local IPFS node URL (development, default: `http://localhost:5001`) |
| `BERYX_API_TOKEN` | [Beryx](https://docs.zondax.ch/beryx) JWT token (Filecoin chain queries) |
| `FILECOIN_RPC_URL` | Filecoin RPC endpoint (Beryx or Glif) |
| `FILECOIN_DATA_URL` | Beryx data API URL |
| `FLOW_ACCESS_NODE` | Flow REST API (public, no auth: `https://rest-testnet.onflow.org`) |
| `FLOW_ACCOUNT_ADDRESS` | Flow account address (`0xf8a2fcf3389475a1` on testnet) |
| `FLOW_PRIVATE_KEY` | Flow ECDSA_P256 private key (hex) |
| `CONTRACT_ADDRESS` | Deployed ProofOfSkill contract address |
| `LIT_NETWORK` | Lit Protocol network (`naga` for v1 Naga SDK) |
| `LIT_API_KEY` | Lit Protocol API key (planned) |

See [`.env.example`](.env.example) for the full template with inline documentation.

---

## Live Deployments

| Service | Network | Address / URL |
|---|---|---|
| ProofOfSkill contract | Flow Testnet | [`0xf8a2fcf3389475a1`](https://testnet.flowscan.io/account/0xf8a2fcf3389475a1) |
| Lighthouse gateway | IPFS/Filecoin | `https://gateway.lighthouse.storage/ipfs/{cid}` |
| Beryx RPC | Filecoin Mainnet | `https://api.zondax.ch/fil/node/mainnet/rpc/v1` |
| Flow REST API | Testnet | `https://rest-testnet.onflow.org/v1` (public, no auth) |
| Impulse AI | Cloud | `https://api.impulselabs.ai/api/chat` (SSE streaming) |

---

## How to Use

1. Open the web app at `http://localhost:3000`
2. Upload a file, paste code, or enter a GitHub repo link
3. Click **"Analyze"**
4. Review the AI feedback (score, issues, suggestions)
5. Click **"Mint Proof"**
6. Receive:
   - IPFS link to the full evaluation report
   - Blockchain transaction ID and on-chain credential

### Verification

Anyone can independently verify a credential:

- Retrieve the report via its **CID** from IPFS
- Compare the CID hash against the **on-chain record**

---

## Example Runs

### Example 1 — Code Submission

```zsh
$ curl -X POST http://localhost:8080/api/submit \
    -F "file=@main.go" \
    -H "Content-Type: multipart/form-data"

{
  "score": 82,
  "issues": [
    { "type": "security", "message": "Unvalidated user input at line 45" },
    { "type": "quality",  "message": "Unused variable 'tmp' at line 12" }
  ],
  "suggestions": [
    "Add input sanitization before DB query",
    "Remove dead code to improve maintainability"
  ],
  "cid": "QmXoypizjW3WknFiJnKLwHCnL72vedxjQkDDP1mXWo6uco",
  "tx_hash": "0xabc123...def456"
}
```

### Example 2 — GitHub Repo Analysis

```zsh
$ curl -X POST http://localhost:8080/api/submit \
    -H "Content-Type: application/json" \
    -d '{"repo": "https://github.com/user/project"}'

{
  "score": 91,
  "issues": [],
  "suggestions": [
    "Consider adding rate limiting to API endpoints"
  ],
  "cid": "QmT5NvUtoM5nWFfrQdVrFtvGfKFmG7AHE8P34isapyhCxX",
  "tx_hash": "0x789abc...123def"
}
```

### Example 3 — Verify a Credential

```zsh
$ curl http://localhost:8080/api/verify/QmXoypizjW3WknFiJnKLwHCnL72vedxjQkDDP1mXWo6uco

{
  "valid": true,
  "on_chain_hash": "0xabc123...def456",
  "timestamp": "2025-07-14T10:30:00Z",
  "score": 82
}
```

---

## Security

### Threat Model

```mermaid
flowchart TD
    A[User Input] -->|Malicious Payload| B[AI Engine]
    B -->|Manipulated Output| C[False Proof]
    C --> D[Blockchain Record]
```

### Risks & Mitigations

| Risk | Description | Mitigation |
|---|---|---|
| AI Poisoning | Malicious inputs trick the model | Input sanitization + multi-model validation |
| Data Exposure | IPFS content is public by default | Encrypt reports with [Lit Protocol v1 Naga SDK](https://developer.litprotocol.com/) |
| Smart Contract Bugs | Exploits in mint logic | Minimal contract surface + formal audits |
| Replay Attacks | Reusing old proofs | Unique submission IDs + timestamps |

### Reporting Vulnerabilities

See [SECURITY.md](SECURITY.md) for the responsible disclosure policy.

---

## CI/CD & Automations

The project uses GitHub Actions for continuous integration:

| Workflow | Trigger | Purpose |
|---|---|---|
| `ci.yml` | Push / PR to `main` | Lint, test, and build backend + frontend |
| `contracts.yml` | Changes in `contracts/` | Compile and test smart contracts |
| `security.yml` | Weekly schedule | Dependency audit + SAST scan |

### Recommended Automation Setup

```yaml
# .github/workflows/ci.yml
name: CI
on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  backend:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.21'
      - run: go test ./...

  frontend:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: '18'
      - run: cd frontend && npm ci && npm run build

  ai:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-python@v5
        with:
          python-version: '3.10'
      - run: cd ai && pip install -r requirements.txt && pytest
```

---

## Use Cases

- **Freelance Verification** — Prove delivered work quality to clients
- **Developer Portfolios** — Back project claims with verifiable AI scores
- **Security Audit Certification** — Immutable proof of audit completion
- **Academic Credentialing** — Tamper-proof assignment and thesis evaluations
- **OSINT & Cyber Investigations** — Verifiable intelligence validation

---

## Future Enhancements

- Autonomous AI agents for continuous code monitoring
- DAO-based community validation layer
- Encrypted private proofs (zero-knowledge)
- On-chain reputation scoring system
- Integration with hiring platforms (LinkedIn, Upwork)

---

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/your-feature`)
3. Commit with clear messages (see commit conventions below)
4. Push and open a Pull Request

### Commit Conventions

- Update `.gitignore` before committing new module types
- Use a short, precise commit title
- Include a well-explained commit body describing the effect on functionality and performance

---

## License

This project is licensed under the [MIT License](LICENSE).

Copyright (c) 2026 Joel Amos
