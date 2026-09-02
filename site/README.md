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
| `getting-started.html` | Install and first deploy |
| `ship.html` | Ship workflow |
| `commands.html` | CLI reference |
| `providers.html` | Providers + common fixes |
| `assets/styles.css` | Theme |
| `assets/site.js` | Sticky header + reveal motion |
| `vercel.json` | Clean URLs + basic headers |
