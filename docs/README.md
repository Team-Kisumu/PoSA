# PoSA Documentation

This directory contains detailed documentation for each component of the Proof-of-Skill AI system.

## Tech Stack

| Technology | Role | Status | Verified |
|---|---|---|---|
| [Impulse AI](https://docs.impulselabs.ai/) | AI-powered code analysis via SSE streaming API | Integrated | Live — scores code 15-100 |
| [Lighthouse.storage](https://docs.lighthouse.storage/) | Primary Filecoin/IPFS pinning (upload + retrieve) | Integrated | Live — roundtrip verified |
| [Filecoin (Beryx)](https://docs.zondax.ch/beryx) | Chain queries via authenticated RPC | Integrated | Live — chain ID 314 |
| [go-synapse](https://github.com/data-preservation-programs/go-synapse) | Native Filecoin PDP storage (fallback) | Integrated | SDK ready, 0 providers |
| [Flow](https://developers.flow.com/) | Blockchain proof anchoring (Cadence) | Deployed | Live on testnet — credential minted |
| [Lit Protocol v1 Naga SDK](https://developer.litprotocol.com/) | Report encryption and access control | Planned | — |

## Component Docs

| Document | Component | Status |
|---|---|---|
| [Backend API](backend.md) | Go HTTP server, routing, validation, middleware | Implemented |
| [AI Engine](ai-engine.md) | Pattern-based analysis (25 langs) + Impulse AI | Implemented |
| [Storage](storage.md) | Lighthouse + Beryx RPC + go-synapse + local IPFS | Implemented |
| [Blockchain](blockchain.md) | Flow ProofOfSkill contract — deployed to testnet | Deployed |
| [Frontend](frontend.md) | React/Next.js UI | Planned |

## Infrastructure Docs

| Document | Scope |
|---|---|
| [Testing](testing.md) | Test strategy, scripts, coverage, e2e tests |
| [CI/CD](ci-cd.md) | GitHub Actions workflows, hooks, automation |

## Live Deployments

| Service | Network | Address / URL |
|---|---|---|
| ProofOfSkill contract | Flow Testnet | [`0xf8a2fcf3389475a1`](https://testnet.flowscan.io/account/0xf8a2fcf3389475a1) |
| Flow account | Flow Testnet | `0xf8a2fcf3389475a1` (ECDSA_P256 / SHA3_256) |
| Lighthouse gateway | IPFS/Filecoin | `https://gateway.lighthouse.storage/ipfs/{cid}` |
| Beryx RPC | Filecoin Mainnet | `https://api.zondax.ch/fil/node/mainnet/rpc/v1` |
| Flow REST API | Testnet | `https://rest-testnet.onflow.org/v1` (public, no auth) |

## Quick Links

- [Main README](../README.md)
- [Environment Config](../.env.example)
- [Security Policy](../SECURITY.md)
- [License](../LICENSE)
