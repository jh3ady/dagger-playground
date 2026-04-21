import { dag, object, func } from "@dagger.io/dagger"

type Cell = { major: string; base: string; tag: string }
type Result = { cell: Cell; ok: boolean; durationMs: number; message: string }

const MATRIX: Cell[] = [
  { major: "15", base: "alpine", tag: "postgres:15.17-alpine" },
  { major: "15", base: "debian", tag: "postgres:15.17-bookworm" },
  { major: "16", base: "alpine", tag: "postgres:16.13-alpine" },
  { major: "16", base: "debian", tag: "postgres:16.13-bookworm" },
  { major: "17", base: "alpine", tag: "postgres:17.9-alpine" },
  { major: "17", base: "debian", tag: "postgres:17.9-bookworm" },
]

@object()
export class MatrixPgCompatTs {
  @func()
  async ping(dsn: string): Promise<string> {
    return dag
      .container()
      .from("alpine:3.21")
      .withExec(["apk", "add", "--no-cache", "postgresql17-client"])
      .withEnvVariable("PGCONNECT_TIMEOUT", "30")
      .withExec(["sh", "-c", `psql "${dsn}" -At -c 'SELECT version();'`])
      .stdout()
  }

  @func()
  async testMatrix(): Promise<string> {
    const results: Result[] = await Promise.all(
      MATRIX.map((cell) => this.runCell(cell)),
    )
    const table = formatResults(results)
    const failed = results.filter((r) => !r.ok).length
    if (failed > 0) {
      throw new Error(`${failed} cell(s) failed\n${table}`)
    }
    return table
  }

  private async runCell(cell: Cell): Promise<Result> {
    const start = Date.now()
    try {
      const pg = dag
        .container()
        .from(cell.tag)
        .withEnvVariable("POSTGRES_HOST_AUTH_METHOD", "trust")
        .withExposedPort(5432)
        .asService()

      const output = await dag
        .container()
        .from("alpine:3.21")
        .withExec(["apk", "add", "--no-cache", "postgresql17-client"])
        .withServiceBinding("pg", pg)
        .withExec([
          "sh",
          "-c",
          `for i in $(seq 1 30); do pg_isready -h pg -U postgres && break; sleep 1; done; ` +
            `psql 'postgres://postgres@pg:5432/postgres' -At -c 'SELECT version();'`,
        ])
        .stdout()

      const needle = `PostgreSQL ${cell.major}.`
      const ok = output.includes(needle)
      return {
        cell,
        ok,
        durationMs: Date.now() - start,
        message: ok ? `matched ${needle}` : `expected ${needle}, got: ${output.trim()}`,
      }
    } catch (err) {
      return {
        cell,
        ok: false,
        durationMs: Date.now() - start,
        message: err instanceof Error ? err.message : String(err),
      }
    }
  }
}

function formatResults(results: Result[]): string {
  const lines: string[] = []
  lines.push("+----------+--------+--------+--------+")
  lines.push("| PG major | Base   | Status | Time   |")
  lines.push("+----------+--------+--------+--------+")
  let ok = 0
  for (const r of results) {
    const status = r.ok ? "OK" : "FAIL"
    if (r.ok) ok++
    const secs = (r.durationMs / 1000).toFixed(1)
    lines.push(
      `| ${r.cell.major.padEnd(8)} | ${r.cell.base.padEnd(6)} | ${status.padEnd(6)} | ${secs.padStart(5)}s |`,
    )
  }
  lines.push("+----------+--------+--------+--------+")
  lines.push(`${ok}/${results.length} passed`)
  for (const r of results) {
    if (!r.ok) lines.push(`FAIL ${r.cell.major}/${r.cell.base}: ${r.message}`)
  }
  return lines.join("\n")
}
