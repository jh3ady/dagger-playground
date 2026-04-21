# dagger-playground

[![CI](https://github.com/jh3ady/dagger-playground/actions/workflows/ci.yml/badge.svg)](https://github.com/jh3ady/dagger-playground/actions/workflows/ci.yml)
[![OpenSSF Scorecard](https://api.securityscorecards.dev/projects/github.com/jh3ady/dagger-playground/badge)](https://securityscorecards.dev/viewer/?uri=github.com/jh3ady/dagger-playground)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

I'd read enough about Dagger to be curious, and I wanted to see what it actually feels like once you're past the "hello world" in the docs. So I picked a concrete problem and built it side by side in all three official SDKs to see where they diverge. This repo is what came out of it.

It's not a library or a starter. It's a journal. I'll add experiments when I find something worth writing down.

## What's here

For now there's one experiment: `matrix-pg-compat`. Six parallel Postgres version checks (15, 16 and 17 on Alpine and Debian bases), orchestrated through Dagger service bindings, done once per SDK.

| #  | Experiment                                              | Status | Gist                                                             |
|----|---------------------------------------------------------|--------|------------------------------------------------------------------|
| 01 | [`matrix-pg-compat`](experiments/01-matrix-pg-compat/)  | done   | 6-cell PG compat matrix (PG 15/16/17 x alpine/debian), parallel. |

Each experiment has its own README and a `FINDINGS.md` where I write up whatever caught me off guard.

## Running it

You need Docker and Dagger. The pinned version is in `.tool-versions` (`brew install dagger` on macOS is the easy path).

```sh
cd experiments/01-matrix-pg-compat/go      && dagger call test-matrix
cd experiments/01-matrix-pg-compat/ts      && dagger call test-matrix
cd experiments/01-matrix-pg-compat/python  && dagger call test-matrix
```

First run is slow because it pulls six Postgres images. After that, Dagger's cache earns its keep.

## A note on the implementations

None of the three SDKs use a native Postgres client here. Each cell shells out to `psql` from an Alpine container. That keeps the three codepaths short and actually comparable, and it puts the spotlight on Dagger itself (matrix, service bindings, parallelism) rather than on SQL driver APIs. A pgx / postgresjs / asyncpg variant would make a good follow-up.

Python is managed with `uv`.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md). If you want to propose a new experiment, open an issue first so we can align on the hypothesis.

## License

[MIT](LICENSE) (c) 2026 Jean-Denis VIDOT.
