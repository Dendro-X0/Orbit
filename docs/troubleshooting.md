# Troubleshooting

Fix the failing step, then re-run or `orbit retry`. Artifacts live under `.orbit/runs/<timestamp>/`.

## Doctor fails

`orbit doctor` reports missing CLIs or auth.

1. Install the provider CLI for your scope (`wrangler`, `vercel`, `fly`, `netlify`)
2. Confirm it is on `PATH`
3. `orbit login <provider>` or complete the provider’s own login
4. Re-run `orbit doctor`

## Login opens a browser you do not want

Use headless / CI paths:

```bash
orbit login cloudflare --guide
orbit login cloudflare --token -
```

## Deploy failed mid-run

Open `failure.json` in the latest run directory, or:

```bash
orbit retry
orbit logs
```

`orbit retry` resumes from the failed step ID unless you pass `--from-step`.

## Status still asks for Vercel after an API-only deploy

Your ship scope may be full-stack, or state may be stale. Run `orbit ship`, choose **API / backend**, pick Cloudflare, and continue. Confirm `.orbit/state.json` `ship.providers` lists only `cloudflare`.

## Worker returns 500 / Cloudflare error 1101

A required binding or secret is often missing. Example: `API_KEY_PEPPER` required at request time.

```bash
orbit secrets
orbit secrets --put API_KEY_PEPPER
curl -sS https://your-worker.example.workers.dev/v1/health
```

Use the health path from `orbit status` recommendations.

## Health check 404

The Worker is up; the path is wrong. Prefer the URL Orbit prints. Many APIs use `/v1/health` rather than `/health`.

## Secrets listed as missing after put

Confirm you put the secret on the same Worker name and environment as `wrangler.toml`. Run `wrangler secret list` from the Worker directory. Redeploy is not required for secret values to take effect on the next request.
