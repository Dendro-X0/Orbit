# Secrets and status

Worker secrets and scoped status keep production deploys honest: Orbit tells you what is missing without dumping provider CLI noise.

## Document secrets in wrangler.toml

Orbit parses comment lines near secrets documentation, then compares names to `wrangler secret list`.

```toml
# Secrets (set via `wrangler secret put`):
# API_KEY_PEPPER, GITHUB_TOKEN, POLAR_WEBHOOK_SECRET
```

Check and set:

```bash
orbit secrets
orbit secrets --put API_KEY_PEPPER
```

Non-interactive put (pipe the value):

```bash
openssl rand -hex 32 | wrangler secret put API_KEY_PEPPER
```

Run that from the Worker directory (for example `apps/api`).

## What status shows

`orbit status` is scoped to the active ship intent:

- Authentication for providers in scope
- Configured targets
- Last successful deploy for this scope (URL, when, run dir)
- Last failed deploy when newer than the last success
- Recommended next: missing secrets, CORS, docs deploy, health URL

Health URLs are detected from API source when possible (for example `app.get("/v1/health", …)`). Fallback is `/health`.

## CORS reminder

If `CORS_ORIGINS` in `wrangler.toml` is localhost-only and an API URL exists, status reminds you to add the production docs origin before shipping the frontend. After you change vars, re-deploy the Worker.
