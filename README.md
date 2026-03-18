# arangoadmin

CLI for managing databases and users in ArangoDB.

## Table of Contents

- [Build & Run](#build--run)
  - [From Source](#from-source)
  - [Docker (local build)](#docker-local-build)
  - [Docker (pre-built image)](#docker-pre-built-image)
- [Global Options](#global-options)
- [Subcommands](#subcommands)
  - [ensure-user](#ensure-user)
  - [ensure-database](#ensure-database)
  - [ensure-grant](#ensure-grant)
- [Quick Reference](#quick-reference)

## Build & Run

### From Source

Requires Go 1.24+.

```sh
git clone https://github.com/dictybase-docker/arangoadmin.git
cd arangoadmin
go build -o arangoadmin ./cmd/arangoadmin/
./arangoadmin [global options] <subcommand> [options]
```

### Docker (local build)

```sh
docker build -t arangoadmin .
docker run --rm arangoadmin <subcommand> [options]
```

The container entrypoint is the binary — pass the subcommand directly after the image name:

```sh
docker run --rm arangoadmin ensure-user \
  --user myuser \
  --password secret \
  --host arangodb.internal
```

For multi-arch builds:

```sh
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  -t arangoadmin .
```

### Docker (pre-built image)

```sh
docker pull ghcr.io/dictybase-docker/arangoadmin:latest
docker run --rm ghcr.io/dictybase-docker/arangoadmin:latest <subcommand> [options]
```

## Global Options

These apply to all subcommands.

| Flag | Default | Env var | Description |
|---|---|---|---|
| `--host` | `arangodb` | `$ARANGODB_SERVICE_HOST` | ArangoDB host address |
| `--port` | `8529` | `$ARANGODB_SERVICE_PORT` | ArangoDB port |
| `--log-level` | `info` | — | Log level (`debug`, `info`, `warn`, `error`) |
| `--log-format` | `json` | — | Log format (`json` or `text`) |
| `--is-secure` | false | — | Connect via TLS |

## Subcommands

### ensure-user

Create a user if it does not exist, or update the user's password according to the configured policy.

```
arangoadmin ensure-user [options]
```

| Flag | Short | Required | Default | Description |
|---|---|---|---|---|
| `--user` | `-u` | yes | — | ArangoDB username |
| `--password` | `--pw` | yes | — | Password for the user |
| `--admin-user` | `--au` | no | `root` | Admin user for authentication |
| `--admin-password` | `--ap` | no | — | Admin password |
| `--password-policy` | — | no | `never` | When to update the password: `never` (set on creation only), `if-provided` (update when a non-empty value is given), `always` (overwrite on every run) |

**Binary:**

```sh
./arangoadmin ensure-user \
  --host localhost \
  --user myuser \
  --password secret \
  --admin-password adminpass \
  --password-policy if-provided
```

**Docker:**

```sh
docker run --rm arangoadmin ensure-user \
  --host localhost \
  --user myuser \
  --password secret \
  --admin-password adminpass \
  --password-policy if-provided
```

---

### ensure-database

Create a database if it does not exist. No-op if the database is already present.

```
arangoadmin ensure-database [options]
```

| Flag | Short | Required | Default | Description |
|---|---|---|---|---|
| `--database` | `--db` | yes | — | Database name |
| `--admin-user` | `--au` | no | `root` | Admin user for authentication |
| `--admin-password` | `--ap` | no | — | Admin password |

**Binary:**

```sh
./arangoadmin ensure-database \
  --host localhost \
  --database mydb \
  --admin-password adminpass
```

**Docker:**

```sh
docker run --rm arangoadmin ensure-database \
  --host localhost \
  --database mydb \
  --admin-password adminpass
```

---

### ensure-grant

Set a user's access level on a database, creating the grant if it does not exist or updating it if it differs.

```
arangoadmin ensure-grant [options]
```

| Flag | Short | Required | Default | Description |
|---|---|---|---|---|
| `--user` | `-u` | yes | — | ArangoDB username |
| `--database` | `--db` | yes | — | Database name |
| `--admin-user` | `--au` | no | `root` | Admin user for authentication |
| `--admin-password` | `--ap` | no | — | Admin password |
| `--grant` | `-g` | no | `rw` | Access level: `rw` (read-write), `ro` (read-only), `none` |

**Binary:**

```sh
./arangoadmin ensure-grant \
  --host localhost \
  --user myuser \
  --database mydb \
  --grant ro \
  --admin-password adminpass
```

**Docker:**

```sh
docker run --rm arangoadmin ensure-grant \
  --host localhost \
  --user myuser \
  --database mydb \
  --grant ro \
  --admin-password adminpass
```

## Quick Reference

| Command | Required flags | Optional flags | Purpose |
|---|---|---|---|
| `ensure-user` | `--user`, `--password` | `--admin-user`, `--admin-password`, `--password-policy` | Create or update a user |
| `ensure-database` | `--database` | `--admin-user`, `--admin-password` | Create a database if missing |
| `ensure-grant` | `--user`, `--database` | `--admin-user`, `--admin-password`, `--grant` | Set user access on a database |
