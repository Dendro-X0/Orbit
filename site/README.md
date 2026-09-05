# Orbit docs site

Static HTML for Orbit. No build step.

## Local preview

```bash
cd site
npx --yes serve .
```

Open the printed URL (paths are root-absolute: `/assets/…`).

## Deploy with Orbit

From the Orbit repo root:

```bash
orbit ship
```

Choose **Static site / docs**, pick Vercel or Netlify, and set the publish directory to `site`.

## Files

| Path | Role |
|------|------|
| `index.html` | Landing |
| `one-stack.html` | Phase 1 HOOK (authorized 2026-09-04) |
| `sample-when-deploy-fails.html` | TRACK sample — pack §6 only (≤20%) |
| `guides.html` | Guide index (all topics) |
| `getting-started.html` | Install and first deploy |
| `ship.html` | Ship workflow |
| `commands.html` | CLI reference |
| `providers.html` | Providers + common fixes |
| `auth.html` | OAuth, tokens, logout |
| `configure-and-deploy.html` | Configure, deploy, retry, wire |
| `secrets.html` | Worker secrets and status |
| `state-and-runs.html` | state.json and run artifacts |
| `troubleshooting.html` | Failures and fixes |
| `assets/styles.css` | Theme |
| `assets/site.js` | Sticky header, reading rail, scroll spy |
| `vercel.json` | Clean URLs + basic headers |

Long-form markdown (auth, configure/deploy, state/runs, troubleshooting) lives in [`../docs/`](../docs/).
