# Findings: matrix-pg-compat

Notes I took while building experiment 01. Started 2026-04-21.

## Observations

- [Go] `dagger init --sdk=go --source=.` scaffolds a working module in one shot. The Go SDK injects `dag` as a package-level variable, so user code doesn't need an explicit import.
- [Go] The idiomatic way to fan out the matrix is goroutines plus `sync.WaitGroup`. There's no first-class "matrix" helper in the SDK. Dagger's lazy graph merges the six branches into one plan automatically.
- [TS] `Promise.all` over the matrix reads exactly as you'd expect. Same end result as Go's goroutine fan-out, Dagger collapses it into one plan.
- [Python] `asyncio.gather` is the idiomatic async primitive here. The SDK's async chaining composes naturally, and the fan-out behaviour matches the Go and TS versions.

## Friction

- [Go] `dag.Container().From(tag).WithEnvVariable(...).WithExposedPort(5432).AsService()` needs the port exposed *before* you call `AsService()`. Forgetting that fails at runtime with a cryptic "port not bound" error, which took me longer than it should have to diagnose.
- [TS] Throwing from inside a `@func()` propagates as a Dagger error with the message preserved. Wrapping the table string in the thrown error makes the CI logs readable. The stack trace is ugly, but the message lands where you need it.
- [Python] `test_matrix` auto-converts to `test-matrix` on the CLI. Same deal as TS `testMatrix`. A per-function CLI-name decorator would save a guess the first time.

## Surprises

- [Go] Alpine 3.20 doesn't ship `postgresql17-client`. That package landed in 3.21. Lots of Dagger examples out there use older Alpine tags, so this bites quickly.
- All three method naming conventions collapse to kebab-case on the CLI: `TestMatrix` (Go), `testMatrix` (TS), `test_matrix` (Python) all become `test-matrix`. Worth knowing when you write the user-facing doc.
- [Python] The uv-based scaffold just works. No `dagger.json` field to select uv: `uv.lock` next to `pyproject.toml` is enough for Dagger 0.20.6 to pick it up.

## What I'd do differently

- Skip the `psql`-in-shell shortcut and try native clients. The current approach keeps the three implementations clean, but it sidesteps the real ergonomic question: what's pgx vs postgresjs vs asyncpg actually like *inside* a Dagger module? A variant experiment would answer that properly.
- Pre-warm the Alpine client base. The first cold run pulls `alpine:3.21` and runs `apk add postgresql17-client` six times (once per cell). A shared client base built once and referenced by each cell would save 30 to 60 seconds on cold runs. Dagger's BuildKit cache handles some of it, but being explicit would be cleaner.
- Factor the ASCII formatter into a shared Dagger module. All three implementations reimplement the same table format with syntactic variations. Pulling it into a module used as a dependency would exercise the "module composition" story I didn't touch here.
- Start with two cells, not six. The first debugging round would have been much faster with just PG 17 alpine and debian. Cold runs with six cells burn two or three minutes each while you're still figuring out how service bindings want to be wired.
