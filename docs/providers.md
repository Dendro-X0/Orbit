# Providers

Orbit detects provider projects in your tree and orchestrates their official CLIs. Detection answers what exists in the repo. Ship **scope** answers what you are shipping in this session.

## Detection table

| Provider | Detection | Typical role |
|----------|-----------|--------------|
| Cloudflare | `wrangler.toml` at repo root and/or `apps/api/wrangler.toml` | Workers API, D1 |
| Vercel | `vercel.json` at root or under `apps/*`, plus Next/Vite markers | Docs / frontend / Next.js |
| Fly.io | `fly.toml` | Apps / containers |
| Netlify | `netlify.toml` | Static sites |

## Scope vs detection

Example: Assess-style monorepo detects Cloudflare and Vercel. You run `orbit ship`, choose **API / backend** → Cloudflare. Then:

- `orbit status` and `orbit configure` talk about Cloudflare only
- Bare `orbit deploy` still deploys **both** unless you pass `--provider cloudflare`

Keep that distinction in mind when status looks “clean” but a full deploy would still touch Vercel.

## Cloudflare

Requires `wrangler` on `PATH`.

### What Orbit does

| Concern | Behavior |
|---------|----------|
| Auth status | `wrangler whoami --json` (plain text fallback) |
| Login | `wrangler login` via ship/login wizards |
| Configure | Create D1 when `database_id` is empty or a `REPLACE*` placeholder; patch `wrangler.toml` |
| Deploy | Phased: whoami → optional `d1 migrations apply` → `wrangler deploy` |
| Secrets | Parse documented names from comments; `secret list` / `secret put` |

### Configure requirement

Cloudflare configure **requires** a D1 `database_name` in `wrangler.toml`. Without a D1 stanza, configure fails even if you only want a Worker with no database. Deploy can skip migration when no D1 name is present; configure cannot.

Minimal shape:

```toml
name = "my-api"
main = "src/worker.ts"
compatibility_date = "2024-11-01"

[[d1_databases]]
binding = "DB"
database_name = "my-db"
database_id = ""  # Orbit fills this on configure when empty
migrations_dir = "migrations"
```

### Secrets documentation

```toml
# Secrets (set via `wrangler secret put`):
# API_KEY_PEPPER, GITHUB_TOKEN
```

Comment blocks whose heading matches `secrets` / `secrets:` / `secrets (` are scanned for `UPPER_SNAKE` names. See [secrets.md](./secrets.md).

### CORS

`CORS_ORIGINS` under `[vars]` feeds status recommendations. Localhost-only values trigger a reminder when an API URL exists. After you change vars, re-deploy the Worker (`[vars]` are not secrets).

## Vercel

Requires the Vercel CLI.

| Concern | Orbit | Provider CLI |
|---------|-------|--------------|
| Detect | `vercel.json`, app folders, Next/Vite | — |
| Configure | Ensure auth; `vercel link --yes` if unlinked; **hint** missing `VITE_*` from `.env.example` | Does not auto-create arbitrary env vars |
| Deploy | Orchestrate phases | `vercel whoami`, `vercel deploy --yes [--prod]` |
| Wire | `orbit wire` / auto-wire in full-stack CF+Vercel deploys | `vercel env rm` / `env add` for `VITE_API_URL` |

Auto-wire runs only when Cloudflare and Vercel are in the **same** deploy provider list, after a Workers URL is extracted from logs, before the Vercel deploy steps.

## Fly.io

Requires `fly` or `flyctl`.

| Orbit | Delegates |
|-------|-----------|
| Detect `fly.toml` | `auth login`, `launch --no-deploy --yes --copy-config` when needed, `status`, `deploy --yes` |

Treated as API / backend in ship project-type selection.

## Netlify

Requires `netlify`.

| Orbit | Delegates |
|-------|-----------|
| Detect `netlify.toml` | `login`, `link`, `status`, `deploy [--prod]`; env hints only |

Treated as web frontend / static site in ship selection.

## Tokens in the OS keychain

When you use `--token` or `--guide`, Orbit stores tokens in the OS keychain under service name `orbit` and maps them to env vars the CLIs understand:

| Provider | Env var |
|----------|---------|
| Cloudflare | `CLOUDFLARE_API_TOKEN` |
| Vercel | `VERCEL_TOKEN` |
| Fly | `FLY_API_TOKEN` |
| Netlify | `NETLIFY_AUTH_TOKEN` |

Browser OAuth via `wrangler login` / `vercel login` uses the provider’s own session store. That is separate from Orbit keychain tokens. See [auth.md](./auth.md).

## What Orbit never replaces

Provider dashboards, billing, DNS, advanced wrangler/vercel flags, and long-tail platform features stay with the provider. Orbit is the portal and the run recorder, not a second control plane.
