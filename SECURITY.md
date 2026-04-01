# Security Policy

## Reporting a Vulnerability

To report a security vulnerability in PoSA, please email the maintainers
directly or [open a draft security advisory](https://github.com/Team-Kisumu/PoSA/security/advisories/new)
so we can coordinate the fix and disclosure.

Please include:
- Description of the vulnerability
- Steps to reproduce
- Affected component (backend, AI engine, smart contract, frontend)
- Potential impact

We will acknowledge receipt within 48 hours and provide an initial assessment
within 5 business days.

---

## Security Architecture

PoSA uses defense-in-depth across four layers:

### Backend (Go)

- 6-step file upload validation: size limit, filename whitelist, content read,
  empty check, MIME sniffing, deep content scan (magic bytes, null bytes, shebangs)
- JSON body size limit (1MB) with strict decoding (DisallowUnknownFields)
- Repo URL validation: HTTPS-only, github.com host, owner/repo path, no traversal
- CID validation: alphanumeric 46-59 chars, injection character blocking
- Security headers: nosniff, DENY, no-referrer, no-store, CSP default-src 'none'
- Panic recovery middleware with structured error responses
- Cryptographic request IDs for tracing

### AI Engine (Python)

- Pydantic request validation with typed fields and length limits
- Pattern-based static analysis with pre-compiled regex (no runtime compilation)
- Score clamped to [0, 100] regardless of input
- No dynamic code execution in the analysis pipeline
- CORS middleware for cross-origin frontend requests

### Smart Contract (Cadence)

- Admin-only minting via resource-based access control
- Duplicate CID prevention (precondition check)
- Score range enforcement (0-100)
- Immutable credentials (struct `let` fields)
- No delete, update, or admin transfer functions
- Formal verification completed with 0 critical/high findings

### Frontend (Next.js)

- Client-side input validation (file size, extension, repo URL format)
- No sensitive data in client-side code
- HTTPS-only API calls

---

## Audit Results

### Smart Contract Formal Verification

**Date:** 2026-04-01
**Contract:** ProofOfSkill.cdc at `0xf8a2fcf3389475a1` (Flow Testnet)

| Severity | Count |
|---|---|
| Critical | 0 |
| High | 0 |
| Medium | 0 |
| Low | 0 |
| Informational | 5 |

Informational findings:
- F-01: No on-chain CID format validation (validated at backend layer)
- F-02: No maximum CID length on-chain (storage fees provide economic limit)
- F-03: Single admin key with no rotation mechanism
- F-04: No pause/emergency stop mechanism
- F-05: Precondition failures revert silently (Cadence design)

Full report: [docs/contract-verification.md](docs/contract-verification.md)

### Input Fuzzing and Adversarial Testing

**Date:** 2026-04-01

| Layer | Test Type | Count | Result |
|---|---|---|---|
| Go Backend | Fuzz tests (5 targets) | ~458K executions | 0 crashes |
| AI Engine | Adversarial tests | 30 cases | All pass |
| Smart Contract | Cadence tests | 10 cases | All pass |

Full report: [docs/security-audit.md](docs/security-audit.md)

---

## Supported Versions

| Version | Supported |
|---|---|
| Current (testnet) | Yes |

---

## Known Limitations

- No rate limiting on API endpoints (recommended for production)
- CORS allows any origin (restrict to production domain)
- Unicode evasion possible against pattern-based analysis (supplemented by Impulse AI)
- Single admin key for smart contract (consider multi-sig for mainnet)
