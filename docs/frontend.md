# Frontend

> **Status:** 📋 Planned (Phase 5)

The frontend provides the user-facing interface for submitting work, viewing AI evaluations, minting proofs, and verifying credentials.

## Technology

- **Framework:** React / Next.js
- **Runtime:** Node.js 18+
- **Port:** `:3000`

## Planned Directory Structure

```zsh
frontend/
├── src/
│   ├── app/                 # Next.js app router pages
│   ├── components/          # Reusable UI components
│   └── lib/                 # API client, utilities
├── package.json
└── next.config.js
```

## How It Will Work

1. User opens `http://localhost:3000`
2. Upload form accepts file drag-and-drop, code paste, or GitHub repo URL
3. Submission is sent to `POST /api/submit` on the backend
4. AI feedback is displayed in real-time (score, issues, suggestions)
5. User clicks "Mint Proof" to trigger blockchain anchoring
6. Result page shows IPFS link and blockchain transaction hash

## Planned Pages

| Route | Purpose |
|---|---|
| `/` | Upload and submission form |
| `/results/{id}` | AI evaluation results and mint action |
| `/verify/{cid}` | Public credential verification |

## Testing

```bash
# From project root
./scripts/test.sh

# Or directly
cd frontend && npm test
cd frontend && npm run lint
```

## Related Issues

- **#12:** Initialize Next.js project and upload interface
- **#13:** AI feedback display and proof minting UI
- **#14:** Credential verification page
