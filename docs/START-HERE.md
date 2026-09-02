# Start here

Orbit is a deploy portal. It does not replace `wrangler`, `vercel`, `fly`, or `netlify`. Those CLIs own auth, deploy, and provider config. Orbit owns the unified workflow: pick what to ship, log phases, and leave agent-readable failures under `.orbit/runs/`.

## Who this is for

Indie and SMB teams with an API Worker, a static docs site, or both. You want one guided path instead of memorizing four provider CLIs.

## Install

Build from source:

```bash
git clone https://github.com/Dendro-X0/Orbit.git
cd Orbit
make build
./orbit version
```

Put `./orbit` (or `orbit.exe` on Windows) on your `PATH`, or call it by full path.

## First deploy

From your project root:

```bash
orbit ship
```

1. Choose project type: **API / backend**, **Static site / docs**, or **Full-stack**
2. Choose provider(s) for that scope
3. Log in, configure, set secrets, then deploy

Orbit remembers the scope in `.orbit/state.json`. Later `orbit status`, `orbit configure`, and recommendations follow that scope.

## After deploy

```bash
orbit status
orbit secrets          # Cloudflare Workers with documented secrets
orbit open --target api
```

Use [Ship workflow](./ship.md) for the guided flow, [Commands](./commands.md) for the full CLI, and [Troubleshooting](./troubleshooting.md) when a step fails.

## Production checklist

1. `orbit ship` with the intended scope
2. `orbit doctor` for scoped provider CLIs and auth
3. `orbit configure` for D1 / project links
4. `orbit secrets` until required Worker secrets are set
5. Deploy via **Ship** or `orbit deploy --provider …`
6. `orbit status`: live URL, no blocking next steps
7. Update `CORS_ORIGINS` when a production docs URL exists
8. Hit the health route Orbit recommends (often `/v1/health` or `/health`)
