# Commands

Reference for the Orbit CLI. Run `orbit <command> --help` for the live flag list. Version strings and help text come from the binary you built.

## Global

| Flag | Default | Purpose |
|------|---------|---------|
| `--path <dir>` | Auto-detect project root | Run as if started in that directory |

Auto-detect walks upward for `.git`, `go.mod`, `package.json`, or `pnpm-workspace.yaml`.

## Guided workflow

| Command | Purpose |
|---------|---------|
| `orbit` | Same as `ship` in a TTY; otherwise prints help |
| `orbit ship` | Project type → provider(s) → action menu (TTY required) |
| `orbit menu` | Loop: ship, deploy, configure, login, doctor, secrets, status, logs, quit |

## Status and health

| Command | Purpose |
|---------|---------|
| `orbit status` | Scoped auth, config, last deploy, failures, recommended next |
| `orbit doctor` | CLI presence + auth checks; non-zero exit if any check fails |
| `orbit whoami` | Connected accounts for registered providers |
| `orbit version` | `orbit <version> [commit] [(build date)]` |

## Auth

| Command | Purpose |
|---------|---------|
| `orbit login [provider]` | Interactive: confirm, then provider browser OAuth |
| `orbit login --all` | Log in to each detected provider, or all registered if none detected; skips already logged in |
| `orbit login --guide` | Open the provider token page, paste token into the OS keychain |
| `orbit login --token <token>` | Store token directly; use `-` to read from stdin |
| `orbit logout [provider]` | Remove Orbit-stored API tokens from the OS keychain only |

`logout` does **not** run `wrangler logout` / `vercel logout`. Provider CLI sessions can remain until you clear them with the provider tool.

Details: [auth.md](./auth.md).

## Configure and deploy

| Command | Purpose |
|---------|---------|
| `orbit configure` | Interactive setup for active ship scope (or detected providers) |
| `orbit configure --provider <id>` | Single provider |
| `orbit configure --all` | Every detected provider |
| `orbit configure --env <name>` | Target environment (default `production`) |
| `orbit configure --dry-run` | Print planned changes without applying |
| `orbit configure --yes` | Skip interactive prompts where supported |
| `orbit deploy` | Deploy **all detected** providers in stack order |
| `orbit deploy --provider <id>` | Deploy one provider |
| `orbit deploy --env <name>` | Environment (default `production`) |
| `orbit retry` | Resume from the failed step in the latest `failure.json` |
| `orbit retry --from-step <id>` | Resume from a specific step ID |
| `orbit retry --provider <id>` | Fallback provider selection when the failure manifest does not carry one |
| `orbit retry --env <name>` | Environment (default `production`) |

Details: [configure-and-deploy.md](./configure-and-deploy.md).

## Secrets, wiring, URLs, logs

| Command | Purpose |
|---------|---------|
| `orbit secrets` | Compare wrangler.toml-documented secrets to `wrangler secret list` |
| `orbit secrets --provider cloudflare` | Provider (Cloudflare only today) |
| `orbit secrets --put NAME` | Interactive `wrangler secret put NAME` |
| `orbit wire` | Set `VITE_API_URL` on Vercel from the last Workers URL in logs |
| `orbit wire --api-url <url>` | Override the API URL to wire |
| `orbit wire --env <name>` | Environment (default `production`) |
| `orbit open` | Open a URL from the latest run (default `--target api`) |
| `orbit open --target api` | Prefer API URL (`*.workers.dev` / `*.fly.dev`) |
| `orbit open --target docs` | Prefer docs URL (`*.vercel.app` / `*.netlify.app`) |
| `orbit open --target any` | Prefer **API**, else docs |
| `orbit logs` | Print the latest run’s `combined.log` |
| `orbit logs --failed` | Prefer failure details from `failure.json` |

## Stack deploy order

When multiple providers deploy in one run:

1. Cloudflare  
2. Fly.io  
3. Vercel  
4. Netlify  

(An internal `railway` slot exists in ordering code but no Railway provider is registered.)

## Run artifacts (short)

Each run writes `.orbit/runs/<UTC-timestamp>/` and updates `.orbit/latest`. See [state-and-runs.md](./state-and-runs.md).
