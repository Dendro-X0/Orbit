# Auth and login

Orbit is a portal: for interactive use it hands off to the provider CLI OAuth flow. For CI and headless use it can store API tokens in the OS keychain and expose them as environment variables the provider CLIs already understand.

## Check current sessions

```bash
orbit whoami
orbit doctor
orbit status
```

`whoami` queries every registered provider. `doctor` fails the process if any check for detected tools/auth fails. `status` only emphasizes providers in the active ship scope.

## Interactive OAuth (default)

```bash
orbit login cloudflare
orbit login vercel
orbit login --all
```

Default interactive path:

1. Confirm opening browser login when prompted
2. Run the provider command (`wrangler login`, `vercel login`, `fly auth login`, `netlify login`)
3. Verify with WhoAmI

`orbit login --all` walks the detected stack, or all registered providers if nothing is detected. Providers that already have an active CLI session are skipped.

Ship and configure call the same “ensure logged in” helper before work. They only start OAuth for providers without a session; they do not invent sessions from browser cookies alone.

## Manual API token wizard

Use when you cannot complete browser OAuth in this environment:

```bash
orbit login cloudflare --guide
```

Orbit opens the provider’s token creation URL, then reads a hidden paste into the keychain and verifies WhoAmI. A failed WhoAmI deletes the stored token.

## Scripted token

```bash
orbit login cloudflare --token "$CLOUDFLARE_API_TOKEN"
printf '%s' "$CLOUDFLARE_API_TOKEN" | orbit login cloudflare --token -
```

Same keychain + WhoAmI verification as `--guide`.

## Keychain mapping

| Provider | Environment variable |
|----------|----------------------|
| Cloudflare | `CLOUDFLARE_API_TOKEN` |
| Vercel | `VERCEL_TOKEN` |
| Fly.io | `FLY_API_TOKEN` |
| Netlify | `NETLIFY_AUTH_TOKEN` |

Service name in the OS keychain: `orbit`.

## Logout

```bash
orbit logout
orbit logout cloudflare
```

Removes Orbit-stored API tokens from the keychain only. It does **not** revoke provider CLI OAuth sessions. To clear those, use the provider CLI (`wrangler logout`, `vercel logout`, and so on).

## Choosing a method

| Context | Method |
|---------|--------|
| Local laptop, first time | `orbit login <provider>` or ship’s prepare path |
| CI / SSH without browser | `--token` or `--guide` |
| Ship says open browser, you refuse | Cancel, then `--token` / `--guide`, re-run ship |

## Auth failures during configure or deploy

When a step fails with an auth hint (`auth.required`), Orbit may offer interactive login and a retry. If that is not possible:

```bash
orbit login <provider> --token -
orbit retry
```

Or fix the provider CLI session directly, then `orbit retry`.
