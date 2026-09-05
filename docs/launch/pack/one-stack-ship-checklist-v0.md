# One-Stack Ship Checklist Pack — v0

**Status:** Draft  
**Price (CONVERT):** $19 intro / $29 standard  
**Stack (v0 only):** Cloudflare Workers API + Vercel / Pages companion  
**Dogfood fixture:** Assess API + docs (`apps/api` Worker + docs/web on Vercel)  
**As-of:** 2026-09-04  
**TRACK sample (free):** [site/sample-when-deploy-fails.html](../../../site/sample-when-deploy-fails.html) — §6 only · no full-pack trial  
**CLI:** Orbit remains free (GIFT). This pack is the ordered production ritual.

Print or copy this sheet. Check boxes as you go. Fill §7 with your URLs.

---

## 1. What Orbit does (and does not)

Orbit is a **deploy portal**. Official provider CLIs stay the source of truth.

| Layer | Owner |
|-------|--------|
| Browser OAuth, project linking, billing, DNS, advanced flags | `wrangler` / `vercel` (and their dashboards) |
| Guided scope, menus, phased logs, `failure.json`, recommended next | Orbit |

**Orbit does**

- Detect providers from `wrangler.toml` / `vercel.json`
- Ask what you are shipping **right now** (API vs docs vs full-stack)
- Persist scope in `.orbit/state.json` → `ship`
- Orchestrate login → configure → secrets → deploy for that scope
- Write `.orbit/runs/<timestamp>/` with redacted logs and agent-readable failures

**Orbit does not**

- Replace `wrangler` or `vercel`
- Deploy “the whole monorepo” when you only meant the API (unless you choose full-stack or bare `orbit deploy`)
- Own Cloudflare / Vercel ToS, quotas, or billing

**Scope vs detection (critical)**

Detection = what exists in the repo. Ship scope = what this session deploys.

Example (Assess-style): repo detects Cloudflare + Vercel. You choose **API / backend → Cloudflare**. Then `orbit status` / `orbit configure` stay Cloudflare-only. Bare `orbit deploy` (no `--provider`) still deploys **both** — prefer `orbit ship` or `orbit deploy --provider cloudflare`.

---

## 2. Doctor preflight

Run from the **application** repo root (not the Orbit source tree unless you are shipping Orbit’s own `site/`).

```bash
orbit doctor
```

Non-zero exit means `doctor found issues`. Fix before production ship:

| Check | Fix |
|-------|-----|
| `wrangler` / `vercel` missing | Install CLI; confirm same shell `PATH` as Orbit |
| Not logged in | `orbit login cloudflare` / `orbit login vercel`, or provider’s own login |
| Wrong directory | `orbit --path /path/to/monorepo doctor` |

Also confirm config files exist for this stack:

- [ ] `wrangler.toml` (root and/or `apps/api/wrangler.toml`)
- [ ] Companion: `vercel.json` (or Pages project you will link via Vercel CLI)

Headless / CI: `orbit ship` needs a TTY. Script with `orbit deploy --provider …` and token login instead.

---

## 3. Ordered ship ritual

One stack, production order. Do not skip secrets honesty.

### 3.1 Decide scope for this pass

| Pass | Ship choice | Providers |
|------|-------------|-----------|
| **A — API first (recommended)** | API / backend | Cloudflare |
| **B — Docs after API healthy** | Static site / docs (or Web application) | Vercel |
| **C — Both in one run (beta)** | Full-stack | Cloudflare + Vercel |

Default selection stays **API / backend** even when the monorepo looks full-stack. Full-stack is an explicit choice.

### 3.2 Ritual checklist (Pass A)

- [ ] `orbit doctor` clean for Cloudflare
- [ ] `orbit ship` → **API / backend** → Cloudflare
- [ ] Menu: **Ship — prepare and deploy**
- [ ] Confirm browser login if Orbit asks (`wrangler login`)
- [ ] Configure: D1 `database_name` present; Orbit creates/patches `database_id` when empty or `REPLACE*`
- [ ] Set missing Worker secrets when prompted (or cancel → finish with `orbit secrets` before deploy)
- [ ] Deploy finishes; note Workers URL
- [ ] `orbit status` — live URL; read **Recommended next**
- [ ] Health curl (§5)
- [ ] CORS: add production docs origin under `[vars]` `CORS_ORIGINS`, then re-deploy Worker when docs URL exists

### 3.3 Ritual checklist (Pass B — companion)

- [ ] API Pass A healthy
- [ ] `orbit ship` → **Static site / docs** → Vercel
- [ ] Login / link if prompted
- [ ] Deploy; note `*.vercel.app` (or custom) URL
- [ ] If docs were deployed without auto-wire: `orbit wire` (sets `VITE_API_URL` from last Workers URL)
- [ ] Update Worker `CORS_ORIGINS` to include the docs origin; re-deploy Cloudflare

### 3.4 Full-stack note (Pass C)

When Cloudflare and Vercel are in the **same** deploy list, Orbit can auto-wire `VITE_API_URL` after the Workers URL is known and before `vercel deploy`. Treat full-stack as beta until you have run both end-to-end on your fixture.

### 3.5 Already deployed honesty

If this provider label already succeeded, the ship menu becomes **Already deployed**:

- Open live URL / status / secrets
- **Re-deploy (not recommended)** — requires confirmation (deploy path may ask again)

There is no casual second Deploy item. Re-ship is deliberate.

### 3.6 Command cheatsheet (scoped)

```bash
orbit doctor
orbit ship
orbit status
orbit secrets
orbit configure --provider cloudflare
orbit deploy --provider cloudflare    # scoped; not bare deploy
orbit deploy --provider vercel
orbit wire
orbit open --target api
orbit open --target docs
orbit open --target any               # API first, else docs
```

---

## 4. Secrets checklist

Document required names in `wrangler.toml` so Orbit can compare against `wrangler secret list`:

```toml
# Secrets (set via `wrangler secret put`):
# API_KEY_PEPPER, GITHUB_TOKEN
```

Heading must match `secrets` / `secrets:` / `secrets (…`. Orbit collects `UPPER_SNAKE` names.

| Step | Done |
|------|------|
| Comment block lists every runtime-required secret | [ ] |
| `orbit secrets` — no missing names (or only optional gaps you accept) | [ ] |
| `orbit secrets --put NAME` for each gap | [ ] |
| If `secret list` fails, Orbit marks **all** documented secrets missing — fix auth/`--path`, re-check | [ ] |

Setting a secret does **not** require re-deploy; the next request sees it. Missing required bindings often show as Worker 500 / Cloudflare 1101.

Ship prepare-and-deploy: cancel on a secret prompt **aborts deploy** and tells you to finish with `orbit secrets`.

`CORS_ORIGINS` is a `[vars]` value, not a secret — edit `wrangler.toml`, then re-deploy the Worker.

---

## 5. Status and health

```bash
orbit status
```

Status follows ship scope (`ship.providers`), not “everything detected.”

Expect sections: path, detected providers, environment, active scope, auth, configured targets, last success URLs, last failure (if newer), **Recommended next**.

| Signal | Typical recommendation |
|--------|------------------------|
| Documented secrets missing | `orbit secrets` |
| `CORS_ORIGINS` localhost-only while API live | Add production docs origin; re-deploy Worker |
| API live, Vercel detected, Vercel not in scope | Later: `orbit ship` → Static site / docs |
| API URL present | Hit health URL |

**Health URL:** Orbit scans API `src` for a `health` route (often `/v1/health`). If none, falls back to `/health`. Prefer the URL status prints.

```bash
curl -sS https://YOUR_WORKER.workers.dev/v1/health
```

- [ ] Health returns success for production Worker
- [ ] `orbit open --target api` opens the same host you curled

---

## 6. When deploy fails

Do not re-run blind. Read the run, then retry.

```bash
orbit logs --failed
# or open:
#   .orbit/latest  →  .orbit/runs/<timestamp>/failure.json
orbit retry
# or
orbit retry --from-step cloudflare-deploy
```

### Artifacts

| File | When | Use |
|------|------|-----|
| `combined.log` | Always | Human session log (redacted) |
| `steps/*` | Always | One phase stdout/stderr |
| `manifest.json` | End of run | Step list, exit codes |
| `summary.json` | Success only | URLs, duration |
| `failure.json` | Failure only | `failedStep`, `message`, `hint`, `logPaths`, `providerRawTail` |

### Hint → action

| `failure.json` hint | Action |
|---------------------|--------|
| `auth.required` | `orbit login …` then `orbit retry` |
| `cli.missing` | Install CLI; fix `PATH`; retry |
| `secrets.missing` | `orbit secrets --put …` |
| `deploy.failed` | Read step logs; fix provider error; retry |

If retry says the last deploy **succeeded**, use `orbit deploy --provider …` or ship **Re-deploy**, not retry.

**Still stuck:** `orbit version` + `orbit doctor` → isolate by running the provider CLI alone in the Worker/docs directory → open a GitHub issue with redacted `manifest.json` / `failure.json`.

---

## 7. Blank run sheet

| Field | Fixture example (Assess) | Yours |
|-------|--------------------------|-------|
| API project / path | `apps/api` (Workers) | |
| Web / docs project | docs or web app (Vercel) | |
| Worker name (`wrangler.toml`) | | |
| D1 `database_name` | | |
| Documented secrets | e.g. `API_KEY_PEPPER`, … | |
| Workers URL | `https://….workers.dev` | |
| Health path | `/v1/health` or status printout | |
| Frontend / docs URL | `https://….vercel.app` | |
| `CORS_ORIGINS` (prod) | | |
| `VITE_API_URL` wired? | Y/N | |
| Last ship date (API) | | |
| Last ship date (docs) | | |
| Last run dir | `.orbit/runs/…` | |

---

## 8. Disclaimer

Not affiliated with Cloudflare, Inc. or Vercel, Inc. Provider CLIs, dashboards, quotas, and Terms of Service apply. Orbit orchestrates those tools; it does not replace them. The Orbit CLI remains free. This checklist pack is optional documentation of a production ritual for one stack (Workers + Vercel/Pages companion) as of the date above. Commands must match your installed Orbit version (`orbit version`); when in doubt, trust `orbit <command> --help` and the repo docs under `docs/`.
