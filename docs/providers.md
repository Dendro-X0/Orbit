# Providers

Orbit detects and orchestrates official provider CLIs. Detection looks for config files in the project tree.

| Provider | Detection | Typical role |
|----------|-----------|--------------|
| Cloudflare | `wrangler.toml` | Workers API, D1 |
| Vercel | `vercel.json` or Vercel project markers | Static / docs frontend |
| Fly.io | `fly.toml` | Containers / apps |
| Netlify | `netlify.toml` | Static sites |

## Cloudflare

Requires `wrangler` on `PATH`. Orbit uses:

- `wrangler login` / whoami (JSON) for auth status
- configure for D1 and project setup when applicable
- deploy via the Cloudflare provider phases
- `orbit secrets` for Worker secrets listed in `wrangler.toml` comments

Document secrets like this so Orbit can check them:

```toml
# Secrets (set via `wrangler secret put`):
# API_KEY_PEPPER, GITHUB_TOKEN
```

`CORS_ORIGINS` under `[vars]` is read for status recommendations. Localhost-only origins trigger a reminder before you ship a production frontend.

## Vercel

Requires the Vercel CLI. Used for static docs and frontends. On full-stack with Cloudflare, Orbit can wire `VITE_API_URL` from the Worker URL before deploy.

## Fly.io and Netlify

Detected when their config files exist. Use `orbit ship` → choose the matching project type and provider. Prefer the provider CLI docs for advanced flags; Orbit wraps the common path.

## Scope vs detection

Detection answers “what is in this repo?” Scope answers “what am I shipping right now?”

An assess-style monorepo can detect Cloudflare and Vercel while the active scope stays API-only. Status then skips Vercel login and configure until you choose Static site or Full-stack in `orbit ship`.
