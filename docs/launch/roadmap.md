# Orbit Phase 1 — Implementation roadmap

**Band:** One-Stack Ship Checklist (HOOK + CONVERT)  
**Owners:** `docs/launch/*`, `specs/launch/*`, `site/` (HOOK mirror)  
**Strategy lock:** H4 standalone · CONVERT = M-D pack · CLI = GIFT · one stack only  

This roadmap sequences **already-specified** Phase 1 objectives. It does not add product scope.

```text
HANDOFF ATOMIC STEP: Wire Checkout when payment rail ready · OR Stage 6 traffic (quality) · kill review 2026-09-25
PAUSED / CANCELLED:  M-X Orbit Cloud as Phase 1 · multi-provider marketing · Assess hard CTA on HOOK
BLOCKED:             Paid checkout announce (Unlock #6 Checkout unchecked)
CANONICAL OWNER:     docs/launch + site/
STAGE:               0–4 done · 5 partial (HOOK+TRACK+footer) · CONVERT checkout deferred
```

---

## Outcome (Definition of Done)

From [phase1.md](./phase1.md):

| # | Deliverable | Track |
|---|-------------|--------|
| D1 | One-stack happy path documented (fixture narrative) | Docs |
| D2 | HOOK page live | HOOK |
| D3 | TRACK sample live (≤20% of pack) | CONVERT preview |
| D4 | Pack file + checkout live | CONVERT |
| D5 | Cafe Kit FAQ filled | Ops |
| D6 | Soft footer stub → Signet | Sibling |
| D7 | Kill window dated | Unlock |

**Funnel (fixed):** `HOOK → GIFT CLI → sample → CONVERT → soft-no → STOP`

---

## Workstreams

| Stream | Job | Primary files |
|--------|-----|----------------|
| **Unlock** | Gate public announce | [unlock-checklist.md](./unlock-checklist.md) |
| **CONVERT** | Ordered ritual pack | [pack/one-stack-ship-checklist-v0.md](./pack/one-stack-ship-checklist-v0.md), [convert-pack-design.md](../../specs/launch/convert-pack-design.md) |
| **HOOK** | One-stack landing | [hook-landing-design.md](../../specs/launch/hook-landing-design.md), `site/` |
| **TRACK** | Public sample section | One of pack §3 or §6 |
| **Ops** | Cafe answers + traffic quality | [cafe-kit.md](./cafe-kit.md), [traffic.md](./traffic.md) |
| **CLI** | Unchanged GIFT — no Phase 1 feature work required | `docs/ship.md`, engineering docs |

---

## Stages

Private drafts may run before unlock. **Public URLs, checkout announce, and paid traffic wait for signed unlock.**

### Stage 0 — Accept scope — **DONE 2026-09-04**

**Goal:** Unlock gates 1–2, 4, 7 ready to check.

| Step | Action | Proof |
|------|--------|-------|
| 0.1 | Maintainer accepts [phase1.md](./phase1.md) (one-stack only) | Unlock #1 ☑ |
| 0.2 | Name dogfood fixture: **Assess API + docs** | Unlock #2 ☑ · pack header |
| 0.3 | Re-affirm won’t-list: no M-X SaaS · no multi-provider ads · no Assess hard CTA | Unlock #4 ☑ |
| 0.4 | Specs co-read: hook + convert designs | Unlock #7 ☑ |

**Exit:** Met. Handoff next = Stage 1 pack expansion. Unlock still LOCKED (#3, #5, #6 + sign-off open).

### Stage 1 — CONVERT pack draft — **DONE 2026-09-04**

**Goal:** Pack stub → reviewable v0 matching [convert-pack-design.md](../../specs/launch/convert-pack-design.md).

| Step | Action | Source |
|------|--------|--------|
| 1.1 | §1 Orbit vs provider CLIs | `docs/START-HERE.md`, `docs/providers.md` |
| 1.2 | §2 Doctor preflight | `docs/troubleshooting.md`, doctor UX |
| 1.3 | §3 Ordered ship ritual (expand beyond outline) | `docs/ship.md`, `docs/configure-and-deploy.md` |
| 1.4 | §4 Secrets checklist | `docs/secrets.md` |
| 1.5 | §5 Status / health curl | ship + status docs |
| 1.6 | §6 Failure honesty / Recommended next | `docs/state-and-runs.md` |
| 1.7 | §7 Run sheet + §8 Disclaimer | Fixture example row filled |
| 1.8 | Commands match current Orbit (no aspirational flags) | L2 vs `docs/commands.md` |

**Exit:** Met. Pack status `Draft` · design “Draft in repo” ☑ · Unlock #3 (friend co-read) still open in parallel.

### Stage 2 — HOOK private draft — **DONE 2026-09-04**

**Goal:** Landing IA from [hook-landing-design.md](../../specs/launch/hook-landing-design.md) without public announce.

| Step | Action | Notes |
|------|--------|-------|
| 2.1 | Surface: `site/one-stack.html` | Existing site theme; `noindex` |
| 2.2 | Compose: H1 · support · demo CTA · ritual · failure · soft Signet | No Assess CTA · no Aff above fold |
| 2.3 | Demo path → getting-started / doctor-ship | GIFT stays free |
| 2.4 | Soft footer stub → Signet (URL TBD) | Unlock #6 still unsigned |

**Exit:** Met. Private path `/one-stack.html` · Cafe Kit HOOK field set · public announce still blocked.

### Stage 3 — TRACK sample — **DONE 2026-09-04**

**Goal:** One public section only (#3 ship order **or** #6 failure next-steps), ≤20% of pack. No full-pack trial.

| Step | Action | Result |
|------|--------|--------|
| 3.1 | Pick section | **§6** When deploy fails |
| 3.2 | Publish sample + link from HOOK | `/sample-when-deploy-fails.html` · linked from `#honesty` |
| 3.3 | Cafe Kit FAQ complete | Level-0 FAQ + sample URL filled (D5) |

**Exit:** Met. Sample opens without checkout · `noindex` until Unlock #6 · soft-no (no full trial).

### Stage 4 — Unlock sign-off — **DONE 2026-09-04**

**Goal:** Flip [unlock-checklist.md](./unlock-checklist.md) from LOCKED → signed.

| Gate | Decision |
|------|----------|
| #5 | **21** days · paid &lt; **3** → PARK CONVERT · review by 2026-09-25 |
| #6 | ☑ HOOK · ☐ Checkout (deferred) · ☑ Signet footer only |
| Sign-off | Maintainer `go` · 2026-09-04 |

**Exit:** Met. Checkout announce still blocked (not authorized). HOOK/TRACK/Signet footer → Stage 5.

### Stage 5 — Public ship (post-unlock) — **PARTIAL 2026-09-04**

Execute only what #6 authorized:

| If authorized | Ship | Status |
|---------------|------|--------|
| HOOK | `/one-stack.html` indexable · linked from `index.html` | ☑ |
| TRACK | `/sample-when-deploy-fails.html` indexable | ☑ |
| Checkout | Pack + paid path | ☐ deferred |
| Signet footer | Soft stub (URL when sibling live) | ☑ stub |

**Exit:** Authorized subset live. CONVERT checkout waits for payment rail + Unlock #6 Checkout. Kill review 2026-09-25.

### Stage 6 — Traffic (quality only)

Per [traffic.md](./traffic.md): prefer after Signet leaf when possible. No Fiverr. No PH-as-strategy.

**Metrics that matter:** doctor/ship demo completes · sample opens · checkout starts · paid · junk→kill channel.

---

## Dependency graph

```text
Stage 0 (accept + fixture)
    ├─► Stage 1 (pack draft) ──► Stage 3 (sample) ──┐
    └─► Stage 2 (HOOK draft) ───────────────────────┼─► Stage 4 (unlock) ─► Stage 5 (public) ─► Stage 6 (traffic)
                                                    │
Friend co-read (#3) ────────────────────────────────┘
```

CLI engineering is **out of Phase 1 critical path** unless pack L2 finds a docs/CLI mismatch — then fix docs (or tiny honesty fix), not new features.

---

## Proof layers (claim discipline)

| Claim | Minimum proof |
|-------|----------------|
| Pack ready | L1 all 8 sections present · L2 commands match `docs/commands.md` · L3 failure honesty wording |
| HOOK ready | L1 won’t-list · L2 links resolve · L3 mobile · L4 Cafe FAQ match |
| CONVERT live | L4 checkout delivers pack artifact |
| Phase 1 done | DoD boxes in phase1.md + unlock signed |

---

## Explicit non-goals (do not schedule)

- M-X Orbit Cloud / hosted SaaS as Phase 1  
- Multi-provider marketing before one-stack proof  
- Assess hard CTA on HOOK  
- Pack trial / ads in Phase 1  
- “Orbit replaces wrangler/vercel” positioning  
- Suite cart with Assess  

---

## Immediate next (matches handoff)

1. Deploy/site publish so `/one-stack.html` and sample are reachable on the host you use.  
2. **Stage 6** quality traffic when ready (no Fiverr; prefer Signet leaf).  
3. Authorize **Checkout** + wire payment when ready; kill review **2026-09-25** (paid &lt; 3 → PARK CONVERT).  

Update [../handoffs/current-session.md](../handoffs/current-session.md) when a stage exits.
