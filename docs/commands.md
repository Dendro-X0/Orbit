# Commands

Reference for the Orbit CLI. Run `orbit <command> --help` for flags.

Global flag: `--path` sets the project root (default: auto-detect upward from the current directory).

## Guided workflow

| Command | Purpose |
|---------|---------|
| `orbit` / `orbit ship` | Project type → provider(s) → login / configure / secrets / deploy |
| `orbit menu` | Interactive picker for individual commands |

## Status and health

| Command | Purpose |
|---------|---------|
| `orbit status` | Scoped config, auth, last deploy, recommended next steps |
| `orbit doctor` | Provider CLIs on `PATH` and auth health |
| `orbit whoami` | Connected provider accounts |
| `orbit version` | Build version, commit, date |

## Auth

| Command | Purpose |
|---------|---------|
| `orbit login [provider]` | Provider browser OAuth by default |
| `orbit login --all` | Log in to each detected (or registered) provider |
| `orbit login --guide` | Manual API token wizard for CI / headless |
| `orbit login --token <token>` | Store a token (use `-` to read stdin) |
| `orbit logout [provider]` | Remove stored API tokens from the OS keychain |

## Configure and deploy

| Command | Purpose |
|---------|---------|
| `orbit configure` | Interactive setup for the active scope (or detected provider) |
| `orbit configure --provider cloudflare` | Single provider |
| `orbit configure --all` | Every detected provider |
| `orbit configure --dry-run` | Show planned changes without applying |
| `orbit deploy` | Deploy detected providers in stack order |
| `orbit deploy --provider cloudflare` | Single provider |
| `orbit retry` | Resume from `failure.json` of the last failed run |
| `orbit retry --from-step <id>` | Resume from a specific step ID |

## Secrets, wiring, URLs

| Command | Purpose |
|---------|---------|
| `orbit secrets` | Compare wrangler.toml-documented secrets to `wrangler secret list` |
| `orbit secrets --put NAME` | Interactive `wrangler secret put` |
| `orbit wire` | Set `VITE_API_URL` on Vercel from the last Workers deploy |
| `orbit open --target api` | Open last API URL |
| `orbit open --target docs` | Open last docs URL |
| `orbit open --target any` | Prefer docs, else API |
| `orbit logs` | View logs from the last deploy run |

## Run artifacts

Each deploy writes `.orbit/runs/<timestamp>/`:

- `combined.log`: full session output (secrets redacted)
- `steps/*.stdout.log` / `steps/*.stderr.log`: per-phase logs
- `manifest.json`: phases, timings, exit codes
- `summary.json`: final outcome
- `failure.json`: present on error (agent entrypoint for `orbit retry`)
