# matrix-pg-compat

Does our code talk to Postgres 15, 16, and 17 correctly across both Alpine and Debian base images? That's six combinations, all running in parallel against ephemeral Postgres services spun up by Dagger. Then the same thing built three times, once per SDK, so I could feel where they diverge.

## The matrix

| PG major | Alpine                  | Debian                    |
|----------|-------------------------|---------------------------|
| 15       | `postgres:15.17-alpine` | `postgres:15.17-bookworm` |
| 16       | `postgres:16.13-alpine` | `postgres:16.13-bookworm` |
| 17       | `postgres:17.9-alpine`  | `postgres:17.9-bookworm`  |

Each cell boots a disposable Postgres service, waits for it to accept connections, runs `SELECT version()`, and checks that the major version in the reply lines up with the tag.

## Three implementations, same shape

All three modules expose the same two functions: `ping(dsn)` as a standalone version check against any DSN, and `test-matrix` which runs the six cells in parallel and returns an ASCII table.

None of them pull in a native Postgres client. Each cell just calls `psql` from an Alpine container (`alpine:3.21` plus `postgresql17-client`). That keeps the three codepaths tight and comparable. The interesting part is how each language expresses the fan-out:

- Go with goroutines and a `sync.WaitGroup`.
- TypeScript with `Promise.all`.
- Python with `asyncio.gather`.

## Run it

```sh
cd go       && dagger call test-matrix
cd ../ts    && dagger call test-matrix
cd ../python && dagger call test-matrix
```

First run is slow, about 1.2 GB of Postgres images get pulled. After that Dagger's cache does its thing.

A passing run looks like:

```
+----------+--------+--------+--------+
| PG major | Base   | Status | Time   |
+----------+--------+--------+--------+
| 15       | alpine | OK     |  ~3s   |
| 15       | debian | OK     |  ~4s   |
| 16       | alpine | OK     |  ~3s   |
| 16       | debian | OK     |  ~4s   |
| 17       | alpine | OK     |  ~3s   |
| 17       | debian | OK     |  ~4s   |
+----------+--------+--------+--------+
6/6 passed
```

See [FINDINGS.md](./FINDINGS.md) for what I actually learned along the way.
