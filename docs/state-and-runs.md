# State and run artifacts

Orbit keeps project intent and deploy history under `.orbit/` in the project root. Add `.orbit/runs/` to `.gitignore` if you have not already; `state.json` is local project state (usually not committed).

## Directory layout

```text
.orbit/
  state.json          # environment, providers, ship scope
  latest              # relative path to the most recent run dir
  runs/
    2026-09-02T06-35-20Z/
      combined.log
      manifest.json
      summary.json      # success only
      failure.json      # failure only
      steps/
        01-cloudflare-whoami.stdout.log
        01-cloudflare-whoami.stderr.log
        …
```

## `state.json`

```json
{
  "environment": "production",
  "providers": {
    "cloudflare": {
      "targetId": "api",
      "configured": true
    },
    "vercel": {
      "targetId": "docs",
      "configured": false
    }
  },
  "ship": {
    "intentId": "api_backend",
    "label": "API / backend — Cloudflare",
    "providers": ["cloudflare"]
  },
  "provider": "cloudflare",
  "targetId": "api",
  "configured": true
}
```

| Field | Meaning |
|-------|---------|
| `environment` | Default env when commands use `--env` default `production` |
| `providers.<id>.targetId` | Selected target inside that provider (for example `api`, `docs`) |
| `providers.<id>.configured` | Last configure attempt marked OK |
| `ship.intentId` | `api_backend`, `web_frontend`, or `full_stack` |
| `ship.label` | Human label shown in status |
| `ship.providers` | Active ship provider list (status / configure scope) |
| Legacy `provider` / `targetId` / `configured` | Older single-provider fields; migrated into `providers` on load |

Edit this file only if you know what you are doing. Prefer `orbit ship` → Change scope.

## Run identity

Each run directory name is a UTC timestamp: `2006-01-02T15-04-05Z`.  
`.orbit/latest` contains the relative path to that directory so `retry`, `logs`, and `open` can find it.

## `combined.log`

Full session output for the run. Secrets and token-like strings are redacted where the runner applies filters. Start here for human debugging.

## `steps/*`

Per-step stdout and stderr. File names include an index and step ID, for example:

- `cloudflare-whoami`
- `cloudflare-migrate`
- `cloudflare-deploy`
- `wire-vite-api-url`
- `vercel-whoami`
- `vercel-deploy`
- `fly-status`

Use these when `combined.log` is noisy or you need one phase only.

## `manifest.json`

Written at end of run. Includes run id, timestamps, provider label, command, root, overall `ok`, and the step list with timings and exit codes. Agents and humans use it to see what ran.

## `summary.json` (success)

Present only when the run succeeded. Includes:

- `ok`
- `apiUrl` / `docsUrl` / `url` when extracted
- `runDir`
- `duration`
- related metadata the runner records

`orbit open` and status “Deployed” sections rely on this data and on URL parsing from logs.

## `failure.json` (failure)

Present only when a step fails. Includes:

- `failedStep` (step ID)
- `message`
- `hint` (classified when possible: `auth.required`, `cli.missing`, `secrets.missing`, `deploy.failed`, …)
- `logPaths`
- `providerRawTail` (recent provider output)

This file is the entrypoint for `orbit retry` and for automation that should not scrape terminals.

## URL extraction rules

From deploy logs / summaries:

| Kind | Pattern (last match wins) |
|------|---------------------------|
| API | `*.workers.dev`, else `*.fly.dev` |
| Docs | `*.vercel.app`, else `*.netlify.app` |

If open fails with “no Workers URL found”, the run logs may not have contained a recognizable URL (failed deploy, or provider output format changed).

## Retention

Orbit does not automatically prune old runs. Delete under `.orbit/runs/` when disk use grows. Keep `failure.json` from a run you still intend to retry, or rely on `.orbit/latest` pointing at it.

## Related commands

```bash
orbit status
orbit logs
orbit logs --failed
orbit retry
orbit open --target api
```
