# Troubleshooting

Fix the failing step, inspect `.orbit/runs/`, then `orbit retry` or re-run the scoped command. Start with `orbit logs` or `orbit logs --failed`.

## Doctor fails

`orbit doctor` returns a non-zero exit when any check fails (`doctor found issues`).

1. Install the CLI for your scope (`wrangler`, `vercel`, `fly`/`flyctl`, `netlify`)
2. Confirm it is on `PATH` in the same shell you use for Orbit
3. `orbit login <provider>` or the provider’s own login
4. Re-run `orbit doctor`

## Ship refuses to start

| Message / symptom | Fix |
|-------------------|-----|
| Requires an interactive terminal | Use a real TTY, or script with `orbit deploy` + `--token` |
| No deployable providers detected | Add `wrangler.toml` / `vercel.json` / `fly.toml` / `netlify.toml`, or pass `--path` to the repo root |

## Login opens a browser you cannot use

```bash
orbit login cloudflare --guide
orbit login cloudflare --token -
```

See [auth.md](./auth.md).

## Configure fails on Cloudflare without D1

Cloudflare configure requires `database_name` under `[[d1_databases]]`. Add a D1 stanza (even if you create the DB through Orbit), then re-run `orbit configure --provider cloudflare`.

## Deploy failed mid-run

```bash
orbit logs --failed
orbit retry
# or
orbit retry --from-step cloudflare-deploy
```

`failure.json` fields: `failedStep`, `message`, `hint`, `logPaths`, `providerRawTail`.

| Hint (when classified) | Likely action |
|------------------------|---------------|
| `auth.required` | `orbit login …` then `orbit retry` |
| `cli.missing` | Install provider CLI, fix `PATH` |
| `secrets.missing` | `orbit secrets --put …` |
| `deploy.failed` | Read step logs; fix provider error; retry |

If retry says the last deploy succeeded, use `orbit deploy` (or ship re-deploy) instead.

## Bare deploy touched providers you did not want

`orbit deploy` without `--provider` deploys the **full detected stack**. Use:

```bash
orbit deploy --provider cloudflare
```

or deploy from `orbit ship` so the saved scope applies.

## Status still asks for Vercel after an API-only deploy

Ship scope may be full-stack, or you never saved an API-only scope.

1. `orbit ship` → **API / backend** → Cloudflare  
2. Confirm `.orbit/state.json` → `ship.providers` is `["cloudflare"]`  
3. Re-run `orbit status`

## Worker returns 500 / Cloudflare error 1101

Often a missing required binding or secret evaluated at request time (example: `API_KEY_PEPPER`).

```bash
orbit secrets
orbit secrets --put API_KEY_PEPPER
curl -sS https://your-worker.example.workers.dev/v1/health
```

Use the health path from `orbit status` recommendations.

## Health check 404

The Worker is reachable; the path is wrong. Prefer the URL status prints. Many APIs use `/v1/health` rather than `/health`.

## Secrets all show missing after a put

1. Run `wrangler secret list` from the Worker directory (`apps/api` or repo root, matching `wrangler.toml`)
2. Confirm Worker `name` matches
3. Re-run `orbit secrets` with `--path` if needed  
If `secret list` itself fails, Orbit marks every documented secret missing.

## `orbit open` finds no URL

The latest run may have failed before a URL was printed, or log parsing did not see `*.workers.dev` / `*.vercel.app` / etc. Open `summary.json` or the provider dashboard, or re-deploy.

## Logout did not clear provider CLI session

Expected. `orbit logout` clears Orbit keychain tokens only. Use `wrangler logout` / `vercel logout` / … for OAuth sessions.

## Double “re-deploy?” prompts

Ship confirms, then the deploy path may confirm again when a prior success exists. Answer both deliberately, or use `orbit deploy --provider …` when you want a single confirmation path.

## Still stuck

1. `orbit version` and `orbit doctor`  
2. `.orbit/latest` → that run’s `failure.json` + step logs  
3. Re-run the provider CLI command alone in the target directory to isolate Orbit vs provider issues  
4. Open an issue at [Dendro-X0/Orbit](https://github.com/Dendro-X0/Orbit) with `manifest.json` / `failure.json` (redact secrets)
