# Gamaj CLI reference

`gamaj cli` is the on-box management tool. It runs the database migrations and
manages admins, users, and subscription links without the dashboard. Every
command reads the same configuration as the service (`/opt/gamaj/.env`, or
`GAMAJ_ENV_FILE` when set), so it always talks to the same database.

```console
$ gamaj cli --help
Usage: gamaj cli <command> [options]

Commands:
  admin          Manage admins
  user           Manage users
  subscription   Subscription helpers
  migrate        Database migrations
```

The database connection comes from `GAMAJ_DATABASE_URL` (required). Add
`--json` where supported to get machine-readable output for automation.

## admin

```console
$ gamaj cli admin --help
Usage: gamaj cli admin <command> [options]

Commands:
  list                 List admins, with --json for automation
  show <admin>          Show one admin and usage counters
  create <username>     Create an admin
  update <admin>        Update role, status, password, or Telegram ID
  set-password <admin>  Reset password and invalidate older tokens
  enable <admin>        Mark admin active
  disable <admin>       Mark admin disabled
  delete <admin>        Soft delete admin
  usage <admin>         Show admin usage counters
  reset-usage <admin>   Reset usage/created traffic counters
```

### Create the first administrator

```bash
sudo gamaj cli admin create --role full_access
```

Flags for `admin create`:

| Flag | Meaning |
|---|---|
| `<username>` / `--username` / `-u` | Admin username (positional also works) |
| `--role` | `standard`, `sudo`, `seller`, or `full_access` |
| `--password` | Password (interactive prompt when omitted) |
| `--random` | Generate a strong random password |
| `--telegram-id` / `--tg` | Link a Telegram account |
| `--json` | Machine-readable output |

`GAMAJ_ADMIN_PASSWORD` can provide the password for non-interactive runs.

### Examples

```bash
gamaj cli admin list
gamaj cli admin list --json
gamaj cli admin show owner
gamaj cli admin create support --role sudo --random
gamaj cli admin update owner --role full_access
gamaj cli admin set-password owner
gamaj cli admin disable reseller-1
gamaj cli admin enable reseller-1
gamaj cli admin usage owner
gamaj cli admin reset-usage owner
gamaj cli admin delete reseller-1 --yes
```

## user

```console
$ gamaj cli user --help
Usage: gamaj cli user <command> [options]
Commands: list, set-owner
```

```bash
gamaj cli user list
gamaj cli user list --username ali
gamaj cli user list --json
gamaj cli user set-owner ali --admin support
```

## subscription

```console
$ gamaj cli subscription --help
Usage: gamaj cli subscription <command> [options]
Commands: get-link, get-config
```

```bash
# Share link for a user
gamaj cli subscription get-link ali

# Client configuration (for example Clash, sing-box, or raw JSON)
gamaj cli subscription get-config ali --format clash
gamaj cli subscription get-config ali --format singbox
```

## migrate

```console
$ gamaj migrate --help
Usage: gamaj migrate <command>
Commands: up [--to version], status
Downgrades are intentionally unsupported.
```

```bash
gamaj cli migrate status
gamaj cli migrate up
gamaj cli migrate up --to 11
```

The panel service runs pending migrations on start, so `migrate up` is mainly
for scripted installs and for recovering a database you manage by hand.

## Exit codes

| Code | Meaning |
|---|---|
| `0` | Success |
| `1` | Error (message on stderr), including a missing or unreachable database |
