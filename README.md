# Orbit

Deploy to your cloud — login, configure, build, and release with synchronized logs for humans and agents.

## Status

Early scaffold. Provider plugins own stack detection and deployment; the core owns UX, auth shortcuts, phased runs, and structured failure output.

## Install (dev)

```bash
go install ./cmd/orbit
```

## Quick start

```bash
cd your-project
orbit doctor
orbit login cloudflare
orbit configure    # interactive wizard
orbit deploy
orbit logs
```

Use `orbit` with no arguments for an interactive menu (TTY required).

When both Cloudflare and Vercel are in the stack, `orbit deploy` automatically wires `VITE_API_URL` on Vercel from the Workers deploy URL before the Vercel deploy phase.

## Commands

| Command | Purpose |
|---------|---------|
| `orbit` | Interactive menu |
| `orbit login [provider]` | Authenticate (`cloudflare`, `vercel`, …) |
| `orbit login --token <token>` | Store an API token in the OS keychain |
| `orbit logout [provider]` | Remove stored API tokens |
| `orbit whoami` | Show connected accounts |
| `orbit configure --all` | Configure every detected provider |
| `orbit deploy` | Deploy all detected providers (API first, then docs) |
| `orbit deploy --provider cloudflare` | Deploy a single provider |
| `orbit wire` | Set `VITE_API_URL` on Vercel from last Workers deploy |
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
cmd/orbit         CLI entry
internal/cli      Cobra commands
internal/run      Phased runner + log sync
internal/provider   Provider interface + registry
internal/providers  Cloudflare, Vercel, …
```

## License

MIT
