# Check Types Reference

Preboot ships with 7 built-in check types. Every check is registered in the global registry via Go's `init()` mechanism, so all types are available the moment the binary starts.

---

## Common fields

All check types share these config fields (full details in [Configuration Reference](configuration.md)):

```yaml
- name: my-check
  type: <one of the 7 types below>
  severity: blocker | warning | info
  options:
    # type-specific options
  message: "Optional custom failure message"
  fix: "Optional remediation instructions"
```

---

## `command_exists`

**Verifies that a command is available in `$PATH`.**

Equivalent to `which <command>` returning a result.

### Options

| Key | Required | Description |
|---|---|---|
| `command` | Yes | Bare command name (no slashes, no arguments) |

### Validation rules

- Must be a bare name: `go`, `docker`, `psql`
- Must not contain path separators (`/` or `\`)
- Must not be empty

### Failure message

```
command "go" not found in PATH
```

### Example

```yaml
- name: go-installed
  type: command_exists
  severity: blocker
  options:
    command: go
  fix: "Install Go from https://go.dev/dl/"

- name: docker-installed
  type: command_exists
  severity: blocker
  options:
    command: docker
  fix: "Install Docker Desktop from https://docs.docker.com/get-docker/"

- name: make-installed
  type: command_exists
  severity: warning
  options:
    command: make
```

---

## `file_exists`

**Verifies that a file exists at a relative path.**

> **Symlink behavior:** the check uses `os.Stat`, which follows symlinks. A symlink pointing to an existing file passes; a broken symlink (target missing) fails. The check confirms existence only — it does not read file contents.

### Options

| Key | Required | Description |
|---|---|---|
| `path` | Yes | Relative path to the file from the working directory |

### Validation rules

- Must be a relative path (must not start with `/`)
- Must not contain `..` (prevents traversal outside the project)
- Must not start with `~`

### Failure message

```
file ".env" not found
```

### Example

```yaml
- name: env-file-exists
  type: file_exists
  severity: blocker
  options:
    path: .env
  fix: "Run: cp .env.example .env"

- name: ssl-cert-exists
  type: file_exists
  severity: warning
  options:
    path: certs/local.pem
  fix: "Run: make gen-certs"
```

---

## `directory_exists`

**Verifies that a directory exists at a relative path.**

> **Symlink behavior:** the check uses `os.Stat`, which follows symlinks. A symlink pointing to an existing directory passes; a broken symlink fails. The check confirms existence only — it does not read directory contents.

### Options

| Key | Required | Description |
|---|---|---|
| `folder` | Yes | Relative path to the directory from the working directory |

### Validation rules

- Must be a relative path (must not start with `/`)
- Must not contain `..`
- Must not start with `~`

### Failure message

```
directory "vendor" not found
```

### Example

```yaml
- name: vendor-dir-exists
  type: directory_exists
  severity: warning
  options:
    folder: vendor
  fix: "Run: go mod vendor"

- name: data-dir-exists
  type: directory_exists
  severity: info
  options:
    folder: tmp/data
```

---

## `env_exists`

**Verifies that a key is present in the `.env` file.**

The check reads `.env` from the current working directory. It parses the file and looks for the key — the value is not validated, only the key's presence.

### Options

| Key | Required | Description |
|---|---|---|
| `key` | Yes | Environment variable name to look for in `.env` |

### .env parsing rules

- Lines starting with `#` are ignored (comments)
- Blank lines are ignored
- Lines without `=` are ignored
- Inline comments are stripped: `KEY=value # comment` → key is present
- Duplicate keys are allowed (last one wins for parsing, but any occurrence counts as present)

### Failure message

```
key DB_URL not found in .env
```

### Example

```yaml
- name: database-url-set
  type: env_exists
  severity: blocker
  options:
    key: DATABASE_URL
  fix: "Add DATABASE_URL=postgres://localhost:5432/mydb to your .env"

- name: stripe-key-set
  type: env_exists
  severity: blocker
  options:
    key: STRIPE_SECRET_KEY

- name: optional-sentry-dsn
  type: env_exists
  severity: info
  options:
    key: SENTRY_DSN
  message: "Sentry not configured — errors will not be tracked locally"
```

### Notes

- The `.env` file is read once and cached across all `env_exists` checks in a single run.
- If `.env` does not exist, every `env_exists` check will fail (consider pairing with a `file_exists` check for `.env`).

---

## `http_reachable`

**Verifies that an HTTP or HTTPS endpoint returns a 2xx response.**

Performs an HTTP GET to the given URL and checks that the response status code is in the 2xx range (200–299).

### Options

| Key | Required | Default | Description |
|---|---|---|---|
| `address` | Yes | — | Full URL including scheme (`http://` or `https://`) |
| `timeout_ms` | No | Global `timeout_ms` | Per-check timeout in milliseconds |

### Validation rules

- Must be a full URL with `http://` or `https://` scheme
- Must not be empty

### Failure message

```
GET http://localhost:8080/healthz returned status 503
```

or, if unreachable:

```
could not connect: dial tcp [::1]:8080: connection refused
```

### Example

```yaml
- name: api-healthy
  type: http_reachable
  severity: blocker
  options:
    address: "http://localhost:8080/healthz"
    timeout_ms: "3000"
  fix: "Start the API server: go run ./cmd/server"

- name: frontend-running
  type: http_reachable
  severity: info
  options:
    address: "http://localhost:3000"

- name: external-api-reachable
  type: http_reachable
  severity: warning
  options:
    address: "https://api.stripe.com/v1"
    timeout_ms: "5000"
```

### Notes

- Skipped in `--quick` mode.
- Redirects are followed.
- Only the status code matters — response body is ignored.

---

## `tcp_reachable`

**Verifies that a TCP connection can be established to a host:port.**

More primitive than `http_reachable` — checks only that the TCP handshake succeeds, not that any application-layer protocol is working. Useful for databases, Redis, message queues, and other non-HTTP services.

### Options

| Key | Required | Default | Description |
|---|---|---|---|
| `address` | Yes | — | `host:port` string |
| `timeout_ms` | No | Global `timeout_ms` | Per-check timeout in milliseconds |

### Validation rules

- Must be in `host:port` format
- Must not include a URL scheme (`http://`, etc.)
- Port must be parseable as an integer

### Failure message

```
could not connect to localhost:5432: dial tcp [::1]:5432: connection refused
```

### Example

```yaml
- name: postgres-reachable
  type: tcp_reachable
  severity: blocker
  options:
    address: "localhost:5432"
  fix: "Start Postgres: docker compose up db -d"

- name: redis-reachable
  type: tcp_reachable
  severity: blocker
  options:
    address: "localhost:6379"
    timeout_ms: "2000"

- name: rabbitmq-reachable
  type: tcp_reachable
  severity: warning
  options:
    address: "localhost:5672"
```

### Notes

- Skipped in `--quick` mode.
- Does not authenticate or send any data after connecting.

---

## `port_free`

**Verifies that a TCP port is NOT currently in use.**

Use this before starting a service that requires a specific port, so you get a clear error instead of a bind failure buried in service logs.

### Options

| Key | Required | Description |
|---|---|---|
| `port` | Yes | Port number as a string (e.g. `"5432"`) |

### Validation rules

- Must be parseable as an integer
- Must be in range 1–65535

### Failure message

```
port 5432 is already in use
```

### Example

```yaml
- name: postgres-port-free
  type: port_free
  severity: blocker
  options:
    port: "5432"
  message: "Port 5432 is taken — is another Postgres already running?"
  fix: "Stop the conflicting service or change the port in docker-compose.yml"

- name: redis-port-free
  type: port_free
  severity: blocker
  options:
    port: "6379"

- name: dev-server-port-free
  type: port_free
  severity: warning
  options:
    port: "3000"
```

### Notes

- The check binds `127.0.0.1:<port>` (loopback only). A service that listens exclusively on a non-loopback interface (e.g. `0.0.0.0` on a remote host, or a container network) may not be detected — the port can appear free even when the service is running.
- Contrast with `tcp_reachable`: `port_free` checks that nothing is listening (port available), whereas `tcp_reachable` checks that something is listening (service up).

---

## Check type quick-reference

| Type | Checks | Network? | Skipped with `--quick`? |
|---|---|---|---|
| `command_exists` | Command in `$PATH` | No | No |
| `file_exists` | File on disk | No | No |
| `directory_exists` | Directory on disk | No | No |
| `env_exists` | Key in `.env` | No | No |
| `http_reachable` | HTTP 2xx response | Yes | Yes |
| `tcp_reachable` | TCP connection | Yes | Yes |
| `port_free` | Port not in use | Local only | No |
