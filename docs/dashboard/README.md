# Gamaj dashboard

The dashboard is a React single-page app (Vite + Chakra UI + TanStack Query).
It is built into static files and embedded into the `gamaj-server` binary, so a
production install never serves the dashboard from a separate process.

## Requirements

- Node.js 20 LTS with npm (the same version CI uses)

## Install dependencies

```bash
cd dashboard
npm ci
```

## Point the dev server at a backend

Copy `example.env` to `.env` and set the API base:

```dotenv
VITE_BASE_API=http://127.0.0.1:616/api/
```

| Variable | Meaning |
|---|---|
| `VITE_BASE_API` | Base URL of the panel API used by the dev server |
| `VITE_ONLINE_ACTIVE_WINDOW_SECONDS` | How long a user counts as online (default: server value) |

Vite's own variables (`VITE_*`) are the only environment names the dashboard
reads; they are allowed by the naming guard because they belong to the bundler.

## Develop

```bash
npm run dev        # dev server with hot reload
npm test           # unit tests
npm run lint       # lint
```

## Production build

```bash
VITE_BASE_API=/api/ npm run build
cp build/index.html build/404.html
```

`scripts/build_dashboard.sh` runs the tutorials build and then this same build,
which is what the release pipeline uses. `scripts/build_binary.sh` copies
`dashboard/build` into `internal/gateway/static/dashboard/build` and embeds it.

## Layout

| Path | Content |
|---|---|
| `src/pages` | Screens (users, nodes, hosts, settings, maintenance, …) |
| `src/components` | Reusable UI, forms, tables, dialogs |
| `src/contexts` | API clients, session, admin and node state |
| `public/statics/locales` | Translations: `en.json`, `fa.json`, `ru.json`, `zh.json` |
| `public/statics` | Static assets shipped with the build |

When you add a user-facing string, add it to all four locale files with the same
key so no screen falls back to a raw key.

## Tests

```bash
npm test           # vitest
npm run lint       # eslint
```

Both run in the release workflow before the binaries are built, and the
end-to-end install test (`scripts/tests/e2e-binary-install.sh`) checks that the
built dashboard is served on port 616.
