"""Dagger module for experiment 01-matrix-pg-compat (Python SDK)."""

import asyncio
import time
from dataclasses import dataclass

from dagger import dag, function, object_type


MATRIX: list[tuple[str, str, str]] = [
    ("15", "alpine", "postgres:15.17-alpine"),
    ("15", "debian", "postgres:15.17-bookworm"),
    ("16", "alpine", "postgres:16.13-alpine"),
    ("16", "debian", "postgres:16.13-bookworm"),
    ("17", "alpine", "postgres:17.9-alpine"),
    ("17", "debian", "postgres:17.9-bookworm"),
]


@dataclass
class CellResult:
    major: str
    base: str
    ok: bool
    duration_s: float
    message: str


@object_type
class MatrixPgCompatPython:
    @function
    async def ping(self, dsn: str) -> str:
        return await (
            dag.container()
            .from_("alpine:3.21")
            .with_exec(["apk", "add", "--no-cache", "postgresql17-client"])
            .with_env_variable("PGCONNECT_TIMEOUT", "30")
            .with_exec(["sh", "-c", f"psql {dsn!r} -At -c 'SELECT version();'"])
            .stdout()
        )

    @function
    async def test_matrix(self) -> str:
        results = await asyncio.gather(
            *(self._run_cell(major, base, tag) for (major, base, tag) in MATRIX)
        )
        table = _format_results(results)
        failed = sum(1 for r in results if not r.ok)
        if failed > 0:
            raise RuntimeError(f"{failed} cell(s) failed\n{table}")
        return table

    async def _run_cell(self, major: str, base: str, tag: str) -> CellResult:
        start = time.monotonic()
        try:
            pg = (
                dag.container()
                .from_(tag)
                .with_env_variable("POSTGRES_HOST_AUTH_METHOD", "trust")
                .with_exposed_port(5432)
                .as_service()
            )

            output = await (
                dag.container()
                .from_("alpine:3.21")
                .with_exec(["apk", "add", "--no-cache", "postgresql17-client"])
                .with_service_binding("pg", pg)
                .with_exec(
                    [
                        "sh",
                        "-c",
                        "for i in $(seq 1 30); do pg_isready -h pg -U postgres && break; sleep 1; done; "
                        "psql 'postgres://postgres@pg:5432/postgres' -At -c 'SELECT version();'",
                    ]
                )
                .stdout()
            )

            needle = f"PostgreSQL {major}."
            ok = needle in output
            return CellResult(
                major=major,
                base=base,
                ok=ok,
                duration_s=time.monotonic() - start,
                message=(f"matched {needle}" if ok else f"expected {needle}, got: {output.strip()}"),
            )
        except Exception as e:  # noqa: BLE001
            return CellResult(
                major=major,
                base=base,
                ok=False,
                duration_s=time.monotonic() - start,
                message=str(e),
            )


def _format_results(results: list[CellResult]) -> str:
    lines: list[str] = []
    lines.append("+----------+--------+--------+--------+")
    lines.append("| PG major | Base   | Status | Time   |")
    lines.append("+----------+--------+--------+--------+")
    ok_count = 0
    for r in results:
        status = "OK" if r.ok else "FAIL"
        if r.ok:
            ok_count += 1
        lines.append(
            f"| {r.major:<8} | {r.base:<6} | {status:<6} | {r.duration_s:>5.1f}s |"
        )
    lines.append("+----------+--------+--------+--------+")
    lines.append(f"{ok_count}/{len(results)} passed")
    for r in results:
        if not r.ok:
            lines.append(f"FAIL {r.major}/{r.base}: {r.message}")
    return "\n".join(lines)
