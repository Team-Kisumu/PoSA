# CI/CD & Automation

This document covers the GitHub Actions workflows, git hooks, and automation scripts that enforce code quality and consistency across the project.

## GitHub Actions Workflows

All workflows live in `.github/workflows/`.

| Workflow | File | Trigger | Purpose |
|---|---|---|---|
| CI/CD | `ci-cd.yml` | Push/PR to `main`/`versions` | Lint, test, build all stacks |
| Code Quality | `quality.yml` | Push/PR to `main`/`versions` | Format checks, linting, type checking |
| Security Scan | `security.yml` | Weekly + push to `main` | Dependency audit, CodeQL SAST, secret detection |
| Issue Triage | `issues.yml` | Issue opened/edited | Auto-labels by title keywords |
| Milestone Tracker | `milestones.yml` | Issue state change | Posts progress bar on milestoned issues |
| Auto Assign | `auto-assign.yml` | PR/issue opened | Assigns to opener, labels PRs by file paths |

### CI/CD Pipeline

Runs on every push and PR. Four parallel jobs:

1. **Backend** — `go mod download` → `go vet` → `go test -race -coverprofile` → `go build`
2. **AI Engine** — `pip install` → `flake8` → `pytest`
3. **Frontend** — `npm ci` → `npm run lint` → `npm run build`
4. **Contracts** — Syntax validation (conditional on contract changes)

### Security Scan

Runs weekly (Monday 06:00 UTC) and on push to `main`:

- **Dependency audit:** `govulncheck` (Go), `pip-audit` (Python), `npm audit` (Node)
- **CodeQL SAST:** Static analysis for Go, JavaScript, Python
- **Secret detection:** TruffleHog scans for leaked credentials

### Code Quality

Enforces formatting and linting standards:

| Stack | Format | Lint | Type Check |
|---|---|---|---|
| Go | `gofmt` | `golangci-lint` | — |
| Python | `black` | `flake8` | `mypy` |
| Frontend | `prettier` | `eslint` | — |

## Git Hooks

Located in `scripts/hooks/`. Install with:

```bash
./scripts/setup-hooks.sh
```

| Hook | When | What it does |
|---|---|---|
| `pre-commit` | Before commit | Auto-formats staged Go/Python/JS files. Blocks `.env` commits. |
| `prepare-commit-msg` | Before commit message | Validates Go build, Python syntax, frontend build. |
| `pre-push` | Before push | Runs `go vet`, `go build`, `flake8`, frontend lint. Blocks on failure. |
| `post-merge` | After merge/pull | Auto-updates deps when lock files change. |

### Hook details

**pre-commit** formats code automatically:
- Go files → `gofmt -w`
- Python files → `black --line-length 120`
- JS/TS files → `prettier --write`
- Blocks any `.env` file from being committed

**pre-push** gates pushes on quality:
- `go vet ./...` and `go build` must pass
- `flake8` must pass for Python
- Frontend lint must pass (if script exists)

**post-merge** keeps deps in sync:
- Detects changes in `go.mod`/`go.sum` → runs `go mod download`
- Detects changes in `requirements.txt` → runs `pip install`
- Detects changes in `package-lock.json` → runs `npm ci`

## Test Scripts

| Script | Usage | Purpose |
|---|---|---|
| `scripts/test.sh` | `./scripts/test.sh` | Run all stacks, summary report |
| `scripts/test-backend.sh` | `./scripts/test-backend.sh -v` | Go vet + tests + coverage |
| `scripts/test-watch.sh` | `./scripts/test-watch.sh backend` | Watch mode, re-run on changes |
| `scripts/setup-hooks.sh` | `./scripts/setup-hooks.sh` | Install git hooks |

## Labels

Custom labels used by workflows:

| Label | Color | Applied by |
|---|---|---|
| `backend` | green | `issues.yml`, `auto-assign.yml` |
| `frontend` | blue | `issues.yml`, `auto-assign.yml` |
| `ai-engine` | purple | `issues.yml`, `auto-assign.yml` |
| `blockchain` | yellow | `issues.yml`, `auto-assign.yml` |
| `security` | red | `issues.yml` |
| `ci/cd` | light blue | `auto-assign.yml` |
| `documentation` | blue | `issues.yml`, `auto-assign.yml` |
| `bug` | red | `issues.yml` |
| `enhancement` | teal | `issues.yml` |
| `triage` | grey | `issues.yml` (default) |
| `needs-info` | peach | `issues.yml` (low-detail issues) |
