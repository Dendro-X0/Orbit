# Ship workflow

`orbit ship` is the guided deploy path. Plain `orbit` with no subcommand does the same thing when stdin is an interactive TTY. In a non-interactive session, bare `orbit` prints help and exits without error.

Ship requires an interactive terminal. If you run it in CI or a pipe, Orbit exits with a message to use an interactive terminal (or use `orbit deploy` / token login for scripting).

## When to use ship vs bare deploy

| Situation | Prefer |
|-----------|--------|
| First deploy, unclear scope, monorepo with API + docs | `orbit ship` |
| Re-run prepare + deploy for the saved scope | `orbit ship` → Ship / Re-deploy |
| Scripted single provider | `orbit deploy --provider cloudflare` |
| Intentionally deploy every detected provider | `orbit deploy` (no `--provider`) |

Bare `orbit deploy` ignores ship scope and deploys the full detected stack in order: Cloudflare → Fly → Vercel → Netlify.

## Entry checks

Before menus appear, ship:

1. Resolves the project root (`--path` or auto-detect)
2. Ensures `.orbit/` and `.orbit/runs/` exist
3. Detects at least one provider (otherwise: no deployable providers)
4. Prints a **Last deploy failed** block when a scoped failure is newer than the last success for that provider

## Step 1: project type

Prompt: **What are you deploying right now?**

Orbit detects components from provider `Detect()` results:

| Provider signal | Treated as |
|-----------------|------------|
| Cloudflare Worker (`wrangler.toml`, often under `apps/api`) | API / backend |
| Fly (`fly.toml`) | API / backend |
| Vercel (`vercel.json`, Next/Vite/static) | Web frontend |
| Netlify (`netlify.toml`) | Web frontend |

| Choice | Intent ID | Providers |
|--------|-----------|-----------|
| **API / backend** | `api_backend` | One backend provider |
| **Static site / docs** (or **Web application** for Next.js) | `web_frontend` | One frontend provider |
| **Full-stack** | `full_stack` | Backend + frontend (or Vercel alone for Next.js full-stack) |

For a monorepo that looks like `apps/api` + `apps/docs`, the profile may **suggest** full-stack, but the default selection remains **API / backend**. Full-stack stays an explicit choice so API-only work does not force a Vercel login.

## Step 2: provider

- One matching provider → Orbit may auto-select it.
- Several options → you pick from a list.
- Full-stack with multiple providers → Orbit builds pairings (for example Cloudflare + Vercel) and may ask: **Deploy to N providers?**

Scope labels look like:

- `API / backend — Cloudflare (production)`
- `Static site / docs — Vercel`
- `Full-stack — Cloudflare + Vercel`
- `Vercel (Next.js full-stack)` when Next.js is the only full-stack target

## Step 3: persist scope

Orbit writes `.orbit/state.json`:

```json
{
  "environment": "production",
  "ship": {
    "intentId": "api_backend",
    "label": "API / backend — Cloudflare",
    "providers": ["cloudflare"]
  },
  "providers": {
    "cloudflare": {
      "targetId": "api",
      "configured": true
    }
  }
}
```

`ship.providers` drives scoped `status` and `configure`. Change scope by choosing **Change scope** in the ship menu, or by running `orbit ship` again.

## Step 4: action menus

### First-time (no successful deploy for this provider label)

Typical items:

1. **Ship — prepare and deploy** (full path)
2. Change scope
3. Log in
4. Configure
5. Set worker secrets
6. Deploy (deploy only, after you already prepared)
7. View status
8. Quit

### Already deployed

Typical items:

1. Open live API / docs URL (when present in the last summary)
2. View status
3. Set worker secrets (when gaps remain)
4. Change scope
5. Log in
6. Configure
7. **Re-deploy (not recommended)** (requires confirmation)
8. Quit

There is no casual second **Deploy** entry on this menu. Re-deploy is explicit.

## Prepare and deploy (full path)

Selecting **Ship — prepare and deploy** or confirmed re-deploy runs:

1. **Login** for each selected provider missing an active CLI session (`ensureProvidersLoggedIn`). Orbit asks before opening browser OAuth.
2. **Configure** for those providers (`configureStack`), then updates state (`configured: true` when successful and not dry-run).
3. **Cloudflare secrets** when in scope: lists documented secrets, offers interactive `wrangler secret put` for each missing name. You can cancel; Orbit then tells you to finish with `orbit secrets` and aborts deploy.
4. **Deploy** via `runDeployForProviders` for the ship provider list.
5. On success from the top-level ship full path, Orbit may offer to open deploy URLs.

You may see **two** re-deploy confirmations: one from the ship menu, and one from the deploy path when a prior success exists.

### Deploy phase order (multi-provider)

1. Cloudflare steps (whoami, optional D1 migrate, deploy)
2. Automatic **wire** of `VITE_API_URL` when both Cloudflare and Vercel are in the same deploy list (after Workers URL is known, before Vercel deploy)
3. Fly (if selected)
4. Vercel / Netlify (if selected)

Exact step IDs appear in `manifest.json` (for example `cloudflare-deploy`, `wire-vite-api-url`, `vercel-deploy`).

## Skipping work Orbit already knows about

- Providers with an active CLI session are not forced through browser login again.
- Secrets already present on the Worker are skipped.
- Prior success for the same provider label switches the menu to Already deployed instead of presenting a blank first-time Ship item.

## Full-stack notes (v0.2.x)

API-only through Cloudflare is the dogfooded production path. Full-stack (Cloudflare + Vercel) works in code, including `VITE_API_URL` wiring, but treat it as beta until you have run both providers end to end on your project.

Next.js-only full-stack can select Vercel alone with label `Vercel (Next.js full-stack)`.

## Related

- [Configure and deploy](./configure-and-deploy.md)
- [Auth](./auth.md)
- [State and run artifacts](./state-and-runs.md)
