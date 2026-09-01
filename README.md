# Orbit

A cross-platform deploy portal — login, configure, deploy, and read logs across cloud providers.

Orbit does not replace official provider tools. It orchestrates them: `wrangler`, `vercel`, and others own auth, deployment, and configuration complexity. Orbit owns the unified UX, phased runs, synchronized logs, and agent-readable failures.

## Install (dev)

```bash
go install ./cmd/orbit
```

## Quick start

```bash
cd your-project
orbit doctor
orbit login --all       # guided OAuth for each provider
orbit configure           # interactive wizard
orbit deploy
orbit logs
```

Use `orbit` with no arguments for an interactive menu (TTY required).

When both Cloudflare and Vercel are in the stack, `orbit deploy` automatically wires `VITE_API_URL` on Vercel from the Workers deploy URL before the Vercel deploy phase.

## Commands

| Command | Purpose |
|---------|---------|
| `orbit` | Interactive menu |
| `orbit login [provider]` | Provider OAuth via official CLI (guided browser flow) |
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
