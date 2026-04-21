# Security Policy

## Supported versions

This is an experimental/educational repository with no released binaries or packages. There is no formal support policy. The `main` branch is what you get.

## Reporting a vulnerability

If you find a security issue, please email `admin@jdevelop.io` directly with the subject prefix `[dagger-playground] security:`. Do not open a public GitHub issue.

Expected first-response time: within 7 days.

## Scope

This repository ships Dagger modules and GitHub Actions workflows. Typical concerns:

- Workflow permission abuse (none use `write`; flag if that changes).
- Pinned image tags disappearing (watched by weekly scheduled CI).
- Supply-chain of the 3 PG clients (watched by Dependabot and CodeQL).

Issues outside these are welcome but may be closed as out-of-scope.
