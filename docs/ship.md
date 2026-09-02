# Ship workflow

`orbit ship` (or plain `orbit` in a TTY) is the guided deploy path. Use it when you want Orbit to pick scope, skip completed steps, and avoid prompting for providers you are not shipping.

## Step 1: project type

| Choice | When to use | Providers |
|--------|-------------|-----------|
| **API / backend** | Workers API, Fly app, or similar | One backend provider |
| **Static site / docs** | Marketing or docs frontend | One frontend provider |
| **Full-stack** | API and docs in one session | Backend + frontend |

When Orbit detects both Cloudflare and Vercel, the default is **API / backend**. Full-stack is an explicit choice so you are not forced into Vercel login for an API-only ship.

## Step 2: provider

Orbit lists providers detected in the repo (`wrangler.toml`, `vercel.json`, `fly.toml`, `netlify.toml`) plus registered options. Pick the provider that matches the project type.

Scope label examples:

- `API / backend — Cloudflare (production)`
- `Static site / docs — Vercel`
- `Full-stack — Cloudflare + Vercel`

## Step 3: actions

After scope is saved, the menu offers login, configure, secrets, deploy, status, and related actions. Orbit checks existing CLI sessions before opening browser OAuth.

If this scope already deployed successfully, you get an **Already deployed** menu: open URL, view status, secrets, or confirm a re-deploy.

## Scope persistence

`.orbit/state.json` stores the active ship intent:

```json
{
  "environment": "production",
  "ship": {
    "intentId": "api_backend",
    "label": "API / backend — Cloudflare",
    "providers": ["cloudflare"]
  }
}
```

`orbit status` and `orbit configure` use this scope. Change scope by running `orbit ship` again and selecting a different project type or provider set.

## Full-stack notes

On full-stack deploy, Orbit typically deploys the API first, then the docs site. When Cloudflare and Vercel are both in scope, Orbit can set `VITE_API_URL` on Vercel from the Worker URL before the Vercel phase (`orbit wire` does this alone).

API-only production is the supported happy path in v0.2.x. Treat full-stack as beta until you have dogfooded both providers end to end.
