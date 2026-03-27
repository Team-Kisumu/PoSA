# PoSA Documentation

This directory contains detailed documentation for each component of the Proof-of-Skill AI system.

## Tech Stack

| Technology | Role | Status |
|---|---|---|
| [Impulse AI](https://docs.impulselabs.ai/) | AI-powered code analysis via SSE streaming API | Integrated |
| [Filecoin (IPFS)](https://docs.filecoin.cloud/) | Decentralized storage via go-synapse + Beryx RPC | Integrated |
| [Lit Protocol v1 Naga SDK](https://developer.litprotocol.com/) | Report encryption and access control | Planned |
| [Flow](https://developers.flow.com/) | Blockchain proof anchoring (Cadence smart contracts) | Planned |

## Component Docs

| Document | Component | Status |
|---|---|---|
| [Backend API](backend.md) | Go HTTP server, routing, validation, middleware | Implemented |
| [AI Engine](ai-engine.md) | Pattern-based analysis (25 langs) + Impulse AI integration | Implemented |
| [Storage](storage.md) | IPFS local + Filecoin via go-synapse + Beryx RPC | Implemented |
| [Blockchain](blockchain.md) | Flow smart contracts, on-chain anchoring | Planned |
| [Frontend](frontend.md) | React/Next.js UI | Planned |

## Infrastructure Docs

| Document | Scope |
|---|---|
| [Testing](testing.md) | Test strategy, scripts, coverage, e2e tests |
| [CI/CD](ci-cd.md) | GitHub Actions workflows, hooks, automation |

## Quick Links

- [Main README](../README.md)
- [Environment Config](../.env.example)
- [Security Policy](../SECURITY.md)
- [License](../LICENSE)
