# Start here

Orbit is a deploy portal for projects that already use official provider CLIs. It does not replace `wrangler`, `vercel`, `fly`, or `netlify`. Those tools own browser OAuth, project linking, and the deploy itself. Orbit owns the guided path: choose what you are shipping, skip finished steps, write phased logs, and leave agent-readable failures under `.orbit/runs/`.

## Who this is for

Indie and SMB teams with a Workers API, a static or Vite docs site, a Fly app, or a mix in one monorepo. You want one interactive flow instead of memorizing four CLIs and remembering which provider you already logged into.

## Prerequisites

- An interactive terminal for `orbit ship` and most wizards
- Provider CLIs on `PATH` for the providers you will use:
  - Cloudflare: `wrangler`
  - Vercel: `vercel`
  - Fly.io: `fly` or `flyctl`
  - Netlify: `netlify`
- Go 1.25+ only if you build from source

## Install

**Release binary (recommended):** download the archive for your OS from
[GitHub Releases](https://github.com/Dendro-X0/Orbit/releases/latest), extract `orbit` (Windows: `orbit.exe`), and put it on your `PATH`.

**Build from source:**

```bash
git clone https://github.com/Dendro-X0/Orbit.git
cd Orbit
make build
./orbit version
```

On Windows the binary is often `orbit.exe`. Put it on your `PATH`, or call it by full path from other projects:

```bash
"/path/to/Orbit/orbit" ship
```

`make build` injects version, commit, and build date via ldflags. Without that, `orbit version` may show a `-dev` version string. Release archives are built the same way via GoReleaser.

## Project root detection

Orbit walks upward from the current directory looking for `.git`, `go.mod`, `package.json`, or `pnpm-workspace.yaml`. Override with:

```bash
orbit --path /path/to/your-project status
```

Every subcommand inherits `--path`.

## First deploy (recommended path)

From your application repo (not the Orbit source tree, unless you are deploying Orbit’s docs site):

```bash
orbit doctor
orbit ship
```

### What `orbit ship` does

1. **Detects** providers from config files (`wrangler.toml`, `vercel.json`, `fly.toml`, `netlify.toml`).
2. Asks **what you are deploying right now**: API / backend, Static site / docs (or Web application for Next.js), or Full-stack.
3. Asks **which provider(s)** match that choice.
4. Saves the scope in `.orbit/state.json` under `ship`.
5. Opens an action menu. Choose **Ship — prepare and deploy** for the full path:
   - Ensure CLI login for selected providers
   - Configure (for example create D1 and patch `database_id`)
   - Prompt for missing Cloudflare Worker secrets
   - Deploy in stack order
   - Offer to open live URLs

If this scope already deployed successfully, the menu becomes an **Already deployed** menu: open URLs, status, secrets, or **Re-deploy (not recommended)** with confirmation.

### After the first success

```bash
orbit status
orbit secrets
orbit open --target api
```

`orbit status` and `orbit configure` follow the saved ship scope. Bare `orbit deploy` without `--provider` deploys the **entire detected stack**, not only the ship scope. Prefer `orbit ship` or `orbit deploy --provider cloudflare` when you mean a single provider.

## Usage map

| Goal | Command / guide |
|------|-----------------|
| Guided first or next deploy | [`ship.md`](./ship.md) · `orbit ship` |
| Full CLI flags | [`commands.md`](./commands.md) |
| Login, tokens, logout | [`auth.md`](./auth.md) |
| Configure, deploy, retry, wire | [`configure-and-deploy.md`](./configure-and-deploy.md) |
| Providers and detection | [`providers.md`](./providers.md) |
| Worker secrets and status | [`secrets.md`](./secrets.md) |
| `.orbit` state and run logs | [`state-and-runs.md`](./state-and-runs.md) |
| Failures and fixes | [`troubleshooting.md`](./troubleshooting.md) |

## Production checklist

1. `orbit ship` with the intended scope (API-only unless you mean full-stack)
2. `orbit doctor` clean for those providers
3. `orbit configure` (Cloudflare needs a D1 `database_name` in `wrangler.toml`)
4. `orbit secrets` until required Worker secrets are set
5. Deploy via **Ship — prepare and deploy**, or `orbit deploy --provider …`
6. `orbit status`: live URL, no blocking recommended next steps
7. Update `CORS_ORIGINS` when a production docs URL exists, then re-deploy the Worker
8. Hit the health URL from status (often `/v1/health`, else `/health`)

## Mental model

| Layer | Owner |
|-------|--------|
| Browser OAuth / API tokens | Provider CLI or Orbit keychain helpers |
| `wrangler deploy`, `vercel deploy`, … | Provider CLI |
| Scope, menus, phased logs, `failure.json`, recommendations | Orbit |
