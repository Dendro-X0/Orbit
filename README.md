# Orbit

A cross-platform deploy portal — login, configure, deploy, and read logs across cloud providers.

Orbit does not replace official provider tools. It orchestrates them: `wrangler`, `vercel`, `fly`, `netlify`, and others own auth, deployment, and configuration complexity. Orbit owns the unified UX, phased runs, synchronized logs, and agent-readable failures.

## Install

**From source (dev):**

```bash
git clone https://github.com/Dendro-X0/Orbit.git
cd Orbit
make build    # versioned binary → ./orbit
make test
```

Or:

```bash
go install ./cmd/orbit
```

## Production checklist (dogfood)

Before calling a release production-ready for your project:

1. `orbit ship` → pick scope (API / frontend / full-stack)
2. `orbit doctor` — CLIs on PATH, auth OK for **scoped** providers only
3. `orbit configure` — D1 / project links (scope-aware)
4. `orbit secrets` — all documented worker secrets set
5. `orbit deploy` or ship → **Ship**
6. `orbit status` — **Live**, no blocking **Recommended next**
7. Update `CORS_ORIGINS` in `wrangler.toml` when docs URL is known
8. Health: `curl <API_URL>/v1/health` (or `/health` depending on your API)

Assess reference: [assess-api DEPLOY.md](https://github.com/Dendro-X0/assess-api/blob/main/DEPLOY.md) (if applicable).

## Quick start

```bash
cd your-project
orbit ship              # pick project type → provider(s) → deploy
orbit status            # scoped to your last ship intent
orbit doctor
orbit login cloudflare  # or use ship for guided OAuth
orbit configure --provider cloudflare
orbit deploy --provider cloudflare
orbit logs
```

Use `orbit` with no arguments to start the guided ship workflow (TTY required). Use `orbit menu` for individual commands.

**Ship flow:** (1) project type — API, static site, or full-stack; (2) provider(s); (3) login / configure / secrets / deploy. Orbit remembers your scope in `.orbit/state.json` and detects prior successful deploys for the same scope.

**Providers:** Cloudflare Workers, Vercel, Fly.io, Netlify (detected via `wrangler.toml`, `vercel.json`, `fly.toml`, `netlify.toml`).

When both Cloudflare and Vercel are in the stack, `orbit deploy` automatically wires `VITE_API_URL` on Vercel from the Workers deploy URL before the Vercel deploy phase.

## Commands

| Command | Purpose |
|---------|---------|
| `orbit` / `orbit ship` | Project type → provider(s) → login / configure / secrets / deploy |
| `orbit menu` | Interactive command picker |
| `orbit status` | Scoped status, deploy history, and next steps for your ship intent |
| `orbit login --all` | Log in to all detected providers in sequence |
| `orbit login --guide` | Manual API token wizard (CI / headless fallback) |
| `orbit login --token <token>` | Store token directly (scripting) |
| `orbit logout [provider]` | Remove stored API tokens |
| `orbit whoami` | Show connected accounts |
| `orbit configure --all` | Configure every detected provider |
| `orbit deploy` | Deploy all detected providers (API first, then docs) |
| `orbit deploy --provider cloudflare` | Deploy a single provider |
| `orbit retry` | Resume from last failed deploy step |
| `orbit open` | Open last deploy URL (`--target api`, `docs`, or `any`) |
| `orbit status` | Project config, auth, last run, suggested next steps |
| `orbit wire` | Set `VITE_API_URL` on Vercel from last Workers deploy |
| `orbit secrets` | Check Worker secrets from wrangler.toml comments |
| `orbit secrets --put NAME` | Set a secret via wrangler |
| `orbit doctor` | Tool + provider health checks |
| `orbit logs` | View last run logs |

## Run artifacts

Each deploy writes to `.orbit/runs/<timestamp>/`:

- `combined.log` — full session output (secrets redacted)
- `steps/*.stdout.log` / `steps/*.stderr.log` — per-phase logs
- `manifest.json` — phases, timings, exit codes
- `summary.json` — final outcome
- `failure.json` — present on error (agent entrypoint)

## Architecture

```
cmd/orbit           CLI entry
internal/cli        Cobra commands + wizards
internal/run        Phased runner + log sync
internal/provider   Provider interface + registry
internal/providers  Cloudflare, Vercel, …
```

## License

MIT
