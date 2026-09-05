# Orbit

Deploy portal for indie and SMB stacks. Orbit orchestrates `wrangler`, `vercel`, `fly`, and `netlify`. Those tools own auth and deploy. Orbit owns the guided ship flow, scoped status, phased logs, and agent-readable failures.

**Docs:** [docs/START-HERE.md](./docs/START-HERE.md) · **Launch Phase 1:** [docs/launch/START-HERE.md](./docs/launch/START-HERE.md) · **Site:** [site/](./site/) (static, deploy with Orbit as Static site / docs)

## Install

**Release binary (recommended):** download the archive for your OS from
[GitHub Releases](https://github.com/Dendro-X0/Orbit/releases), extract `orbit` (or `orbit.exe` on Windows), and put it on your `PATH`.

**Build from source:**

```bash
git clone https://github.com/Dendro-X0/Orbit.git
cd Orbit
make build    # → ./orbit
make test
./orbit version
```

## Quick start

```bash
cd your-project
orbit ship              # project type → provider(s) → deploy
orbit status            # scoped to your last ship intent
orbit doctor
```

Plain `orbit` in a TTY starts the same ship workflow. Use `orbit menu` for individual commands.

**Ship flow:** (1) API, static site, or full-stack; (2) provider(s); (3) login / configure / secrets / deploy. Scope is stored in `.orbit/state.json`. Prior successful deploys for that scope get an Already deployed menu instead of a blind re-deploy.

**Providers:** Cloudflare Workers, Vercel, Fly.io, Netlify (via `wrangler.toml`, `vercel.json`, `fly.toml`, `netlify.toml`).

When Cloudflare and Vercel are both in a full-stack ship, Orbit can set `VITE_API_URL` on Vercel from the Workers URL before the Vercel phase.

## Documentation map

| Doc | Job |
|-----|-----|
| [docs/START-HERE.md](./docs/START-HERE.md) | Install and first deploy |
| [docs/ship.md](./docs/ship.md) | Guided ship workflow, menus, prepare path |
| [docs/commands.md](./docs/commands.md) | Full CLI reference and flags |
| [docs/auth.md](./docs/auth.md) | OAuth, tokens, keychain, logout |
| [docs/configure-and-deploy.md](./docs/configure-and-deploy.md) | Configure, deploy, retry, wire |
| [docs/providers.md](./docs/providers.md) | Detection and Orbit vs provider CLI |
| [docs/secrets.md](./docs/secrets.md) | Worker secrets and status |
| [docs/state-and-runs.md](./docs/state-and-runs.md) | `.orbit` state and run artifacts |
| [docs/troubleshooting.md](./docs/troubleshooting.md) | Failures and fixes |
| [site/](./site/) | Public docs site (HTML) |
| [CHANGELOG.md](./CHANGELOG.md) | Release notes |

**Note:** Bare `orbit deploy` deploys the full detected stack. Prefer `orbit ship` or `orbit deploy --provider …` for scoped deploys.

## Production checklist

1. `orbit ship` with the intended scope
2. `orbit doctor` for scoped providers
3. `orbit configure`
4. `orbit secrets` until required Worker secrets are set
5. Deploy via Ship or `orbit deploy --provider …`
6. `orbit status` shows a live URL without blocking next steps
7. Update `CORS_ORIGINS` when a production docs URL exists
8. Hit the health route Orbit recommends

## Deploy this docs site

```bash
cd Orbit
orbit ship
# Static site / docs → Vercel (or Netlify)
# Project / root directory: site
```

Or point any static host at the `site/` folder (`vercel.json` included).

## Commands (short list)

| Command | Purpose |
|---------|---------|
| `orbit ship` | Guided deploy by project type and provider |
| `orbit status` | Scoped status and recommended next steps |
| `orbit configure` / `orbit deploy --provider …` | Setup and scoped deploy |
| `orbit secrets` / `orbit secrets --put NAME` | Cloudflare Worker secrets |
| `orbit retry` / `orbit logs` / `orbit logs --failed` | Resume failed runs and inspect artifacts |
| `orbit open --target api\|docs\|any` | Open last deploy URL (`any` prefers API) |
| `orbit login` / `orbit wire` / `orbit doctor` | Auth, Vercel `VITE_API_URL`, health |

Full tables: [docs/commands.md](./docs/commands.md).

## Run artifacts

Each deploy writes `.orbit/runs/<timestamp>/` with `combined.log`, per-step logs, `manifest.json`, `summary.json`, and `failure.json` on error.

## Architecture

```
cmd/orbit           CLI entry
internal/cli        Cobra commands + wizards
internal/run        Phased runner + log sync
internal/provider   Provider interface + registry
internal/providers  Cloudflare, Vercel, Fly, Netlify
docs/               Markdown guides
site/               Static docs site
```

## License

MIT
