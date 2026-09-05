# Secrets and status

Cloudflare Worker secrets and scoped status are how Orbit keeps production honest without dumping full provider CLI noise into your terminal.

## Document secrets in `wrangler.toml`

Orbit looks for comment blocks whose heading matches `secrets`, `secrets:`, or `secrets (…`, then collects `UPPER_SNAKE` names (at least three characters after the first letter).

```toml
# Secrets (set via `wrangler secret put`):
# API_KEY_PEPPER, GITHUB_TOKEN, POLAR_WEBHOOK_SECRET, POLAR_CHECKOUT_URL
```

Without that documentation, `orbit secrets` reports that no secrets are documented and shows how to add the comment block.

## Check and set

```bash
orbit secrets
orbit secrets --put API_KEY_PEPPER
```

`--put` runs interactive `wrangler secret put` in the Worker directory (for example `apps/api`).

Non-interactive (from the Worker directory):

```bash
openssl rand -hex 32 | wrangler secret put API_KEY_PEPPER
```

### Behavior details

- Comparison uses `wrangler secret list` for the Worker named in `wrangler.toml`.
- If `secret list` fails (auth, wrong directory, network), Orbit treats **all** documented secrets as missing. Fix auth/`--path`, then re-check.
- Setting a secret does not require a re-deploy; the next request sees the new value.
- Put the secret on the same Worker name and environment you deploy.

## Secrets during `orbit ship`

The prepare-and-deploy path calls into the same missing-secret logic. For each gap you can set the value interactively or cancel. Cancel aborts deploy and prints a reminder to finish with `orbit secrets`.

Status may show **Live with gaps — worker secrets still missing** when the Worker is up but documented secrets are still absent. Required runtime bindings (for example `API_KEY_PEPPER`) can make the Worker return 500 / Cloudflare 1101 until set.

## What `orbit status` shows

Scope resolution order:

1. `ship.providers` from `.orbit/state.json` when present
2. Else provider label from the last successful deploy
3. Else full detected stack

Sections you will see:

- Project path, detected providers, environment, active scope label
- Authentication for scoped providers
- Pending setup / configured targets
- Deployed (this scope): when, duration, API/docs URLs, run dir
- Last deploy failed (when newer than last success for that provider)
- Next command hints
- Recommended next (secrets, CORS, docs deploy, health URL)
- Optional toolkit hints (for example Signet) when relevant

### Recommended next

| Signal | Recommendation |
|--------|----------------|
| Documented Worker secrets missing | `orbit secrets` |
| `CORS_ORIGINS` localhost-only while API is live | Add production docs origin, re-deploy Worker |
| Docs URL known but missing from CORS | Add that origin explicitly |
| API live, Vercel detected, Vercel not in scope | Deploy docs via `orbit ship` → Static site / docs |
| API URL present | Health check URL |

### Health URL detection

Orbit scans `apps/api/src` (or the Cloudflare target `src`) for `.ts` / `.js` files and looks for routes matching `health` (for example `app.get("/v1/health", …)`). If none match, it falls back to `/health`.

```bash
curl -sS https://your-worker.example.workers.dev/v1/health
```

## CORS

`CORS_ORIGINS` lives under `[vars]` in `wrangler.toml` (not secrets):

```toml
[vars]
CORS_ORIGINS = "http://localhost:5173,https://your-docs.example.com"
```

After editing vars, re-deploy the Worker so production picks up the new list.

## Related

- [Providers: Cloudflare](./providers.md#cloudflare)
- [Troubleshooting](./troubleshooting.md)
- [Ship workflow](./ship.md)
