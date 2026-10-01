# Contributing to Gamaj

Thanks for helping. This project has a small set of non-negotiable rules; read
them before you write any code, because a pull request that breaks them will not
be merged.

The full developer reference lives in [`../../SKILL.md`](../../SKILL.md).

## Red lines

1. **Only the Gamaj name.** No reference to any other panel/script project, in
   code, comments, UI text, docs, or assets. The automated guard fails the build
   if such a name appears.
2. **No agent traces.** Do not mention AI assistants, agents, or automation tools
   in code, commit messages, or documentation. Commits must look like normal
   developer work.
3. **One install path: native binary + systemd.** No Docker, no compose files, no
   install-mode switch, and no feature may be disabled based on an install mode.
4. **Every environment variable starts with `GAMAJ_`.** Names owned by the OS or
   third-party tooling are the only exceptions, and they are listed in the guard.
5. **Port 616** stays the panel default.
6. **Docs are self-contained.** Documentation must not send readers to GitHub or
   another site "to read more"; the content belongs in this repository (and in
   `Web/`).

## Workflow

1. Fork or branch from `Asli`.
2. Keep changes focused; one topic per pull request.
3. Run the checks that apply to what you touched:

```bash
cd Panel
gofmt -l .                       # no output
go build ./... && go vet ./internal/... ./cmd/...
go test ./internal/app/system/ ./internal/app/api/ ./internal/gateway/ \
        ./internal/app/backup/ ./internal/app/migrations/ ./cmd/... \
        ./internal/platform/envguard/

cd dashboard
npm ci && npm test && npm run lint
```

4. For installer or script changes, also run `bash -n` on every edited script.

## Commit messages

Short, imperative, and about the change, for example:

```
Fix node registration when the certificate is renewed
Add bandwidth limit to the service editor
```

Do not add tool footers, generated-by lines, or co-author trailers. The commit
author and committer must be your own Git identity.

## Code style

- Go: `gofmt`, `go vet`, error wrapping with `%w`, table-driven tests next to the
  package they cover.
- Dashboard: TypeScript, function components, strings from the locale files.
- Shell: `set -euo pipefail`, `GAMAJ_`-prefixed options, no external dependency
  beyond the usual POSIX tools plus `curl`, `jq`, and `tar`.
- Migrations: append a new numbered migration; never edit an applied one.

## Translations

The dashboard ships English, Persian, Russian, and Chinese. Adding a key means
adding it to `dashboard/public/statics/locales/en.json`, `fa.json`, `ru.json`,
and `zh.json` in the same commit.

## Reporting problems

Include: the Gamaj version (`gamaj status`), the OS and architecture, the exact
command you ran, the output or log lines, and whether the problem happens on a
fresh install. Please remove API keys, tokens, and passwords from anything you
paste.

- Support group: <https://t.me/GamajGP>
- Channel: <https://t.me/TheGamaj>
