# Configure, deploy, retry, and wire

This page covers the non-ship entry points and what each phase does. For the guided menu flow, see [ship.md](./ship.md).

## Configure

```bash
orbit configure
orbit configure --provider cloudflare
orbit configure --all
orbit configure --dry-run
orbit configure --env production
orbit configure --yes
```

### Scope

Without `--all` / `--provider`, configure prefers the active **ship** scope when it is narrower than the full detected stack. That is how API-only projects avoid a Vercel configure prompt.

### Per provider (summary)

| Provider | Configure behavior |
|----------|-------------------|
| Cloudflare | Requires D1 `database_name`. Creates D1 when `database_id` is empty or `REPLACE*`, patches `wrangler.toml` |
| Vercel | Ensures auth; `vercel link --yes` if unlinked; prints hints for missing `VITE_*` from `.env.example` (does not invent env values) |
| Fly | Launch/link style setup via `fly` when needed |
| Netlify | Link / status style setup; env hints only |

On success (and not dry-run), Orbit sets `providers.<id>.configured = true` in `.orbit/state.json`.

If configure hits `auth.required`, Orbit can offer login and retry the configure step.

## Deploy

```bash
orbit deploy --provider cloudflare
orbit deploy --provider vercel --env production
orbit deploy
```

| Invocation | Providers deployed |
|------------|--------------------|
| `--provider <id>` | That provider only |
| No `--provider` | **All detected** providers, in stack order |

Ship scope is **not** applied to bare `orbit deploy`. Use ship’s Deploy / Re-deploy actions, or always pass `--provider`, when you intend a scoped deploy.

### Phases

Orbit builds a step plan, creates `.orbit/runs/<timestamp>/`, streams redacted logs, and writes `manifest.json`. On success it writes `summary.json` (API/docs URLs). On failure it writes `failure.json` and stops.

Example Cloudflare steps: whoami → D1 migrations (when configured) → deploy.  
Example full-stack CF+Vercel: Cloudflare steps → `wire-vite-api-url` → Vercel steps.

### Redeploy confirmation

If a successful deploy already exists for the provider label, deploy asks you to confirm re-deploy to production. Ship may have asked already; answer carefully so you do not double-confirm by accident.

## Retry

```bash
orbit retry
orbit retry --from-step cloudflare-deploy
orbit logs
orbit logs --failed
```

Retry:

1. Reads the latest failure (`.orbit/latest` → `failure.json`)
2. Rebuilds the full step plan for the provider(s)
3. Slices the plan from the failed step ID (or `--from-step`)
4. Writes a **new** run directory

If the latest run succeeded, retry refuses and tells you to use `orbit deploy` instead.

`--provider` on retry is a fallback when the failure record does not carry provider information; the failure manifest is preferred.

## Wire (`VITE_API_URL`)

```bash
orbit wire
orbit wire --api-url https://my-api.example.workers.dev
orbit wire --env production
```

Sets `VITE_API_URL` on the linked Vercel project (remove then add) from:

1. `--api-url` if passed
2. Otherwise the Workers (or similar) URL parsed from the latest successful deploy logs

During a combined Cloudflare + Vercel deploy, Orbit inserts an automatic wire step after the Worker URL is known and before `vercel deploy`. Standalone `orbit wire` covers the case where you deploy providers separately or need to refresh the env var.

Vercel configure **hints** at missing `VITE_*` keys from `.env.example` but does not set them except through wire for `VITE_API_URL`.

## Open and inspect

```bash
orbit open --target api
orbit open --target docs
orbit open --target any    # API first, then docs
orbit status
orbit logs
```

URLs come from the latest run summary / log parse:

- API: last `*.workers.dev`, else `*.fly.dev`
- Docs: last `*.vercel.app`, else `*.netlify.app`

## Recommended sequence (API-only Worker)

```bash
orbit doctor
orbit ship          # API / backend → Cloudflare → Ship — prepare and deploy
orbit status
orbit secrets       # until required secrets are set
curl -sS "$(orbit status | …)/v1/health"   # or the health URL status prints
```

## Recommended sequence (API then docs)

```bash
orbit ship          # API only first
# … production API healthy …
orbit ship          # Static site / docs → Vercel
# or full-stack once you are ready for both
orbit wire          # if docs were deployed without auto-wire
```
