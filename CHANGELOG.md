# Changelog

## Unreleased

### Docs
- Markdown guides under `docs/` (start here, ship, commands, providers, secrets, troubleshooting)
- Static docs site under `site/` for Vercel/Netlify deploy

## 0.2.1 — 2026-09-02

### Fixed
- Health check URL in `orbit status` recommendations detects the route from API source (e.g. `/v1/health` instead of assuming `/health`)

## 0.2.0 — 2026-09-02

### Ship workflow
- Two-step deploy: project type (API / frontend / full-stack) then provider(s)
- Default to API-only when multiple parts detected; full-stack requires explicit choice
- Persist deploy scope in `.orbit/state.json` (`ship` field)
- Detect prior successful deploys; show **Already deployed** and re-deploy confirmation
- Show last failed deploy at `orbit ship` entry with `orbit retry` hint

### Status & recommendations
- `orbit status` scoped to active deploy intent (no spurious Vercel steps for API-only)
- **Recommended next** after deploy: secrets, CORS, docs, health check URL
- Colored TUI output (URLs, commands, warnings)

### Providers
- Cloudflare `whoami` parses JSON (no full wrangler dump in status)
- `orbit configure` follows active ship scope

### Other
- `orbit version` reports build version (use `make build` for release binaries)

## 0.1.x

- Multi-provider deploy: Cloudflare, Vercel, Fly.io, Netlify
- Phased runs, `failure.json`, `orbit retry`, secrets, wire `VITE_API_URL`
