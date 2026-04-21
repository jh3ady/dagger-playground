# Contributing

This is a personal lab, so the bar for "does this fit?" is on the high side. Ideas that test something concrete about Dagger in Go, TypeScript, or Python are welcome. Generic boilerplates, cookbook entries without a hypothesis, or unrelated CI tooling are probably not.

If you're not sure, open an issue first.

## Issue templates

There are three templates in [.github/ISSUE_TEMPLATE/](.github/ISSUE_TEMPLATE/): bug, feature request, and new-experiment proposal. Use them.

For a small fix on an existing experiment, a PR is fine without an issue.

## Running locally

You need Docker and the Dagger CLI pinned in [`.tool-versions`](.tool-versions). Each SDK works the same way:

```sh
cd experiments/01-matrix-pg-compat/<sdk>
dagger develop          # regenerate bindings
dagger call test-matrix
```

Toolchain per SDK: Go 1.23+, Node 22 LTS with npm, Python 3.13 with `uv`.

## Style

- Go: `go fmt ./...` and `goimports` before committing.
- TypeScript: Biome, `npx @biomejs/biome check --write .`.
- Python: Ruff, `uv run ruff check --fix . && uv run ruff format .`.

Commits use [gitmoji](https://gitmoji.dev) plus [Conventional Commits](https://www.conventionalcommits.org). No AI co-author trailers.

## Before you open a PR

Make sure `dagger call test-matrix` runs green locally for every SDK you touched. If you saw something worth noting, update the experiment's `FINDINGS.md`. If the public surface changed, update the README too.

## Adding a new SDK to an experiment

1. Run `dagger init --sdk=<sdk>` in a new subdirectory.
2. Match the existing function shape (`ping`, `test-matrix`).
3. Match the ASCII output format so the CI logs stay readable.
4. Add the SDK to the matrix in `.github/workflows/ci.yml`.
5. Update the experiment's README and FINDINGS.

## Adding a new experiment

Open a new-experiment issue first. Once we agree on the hypothesis, create `experiments/NN-<slug>/` with its own README and FINDINGS, implement in at least one SDK, and update the root README table plus `.github/labeler.yml`.
