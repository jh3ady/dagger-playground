// A Dagger module for experiment 01-matrix-pg-compat (Go SDK).
//
// Exposes:
//   - Ping: run SELECT version() against an arbitrary DSN from a client container.
//   - TestMatrix: run 6 parallel PG service bindings and aggregate results.
package main

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

type MatrixPgCompatGo struct{}

type matrixCell struct {
	major string
	base  string
	tag   string
}

var matrixCells = []matrixCell{
	{"15", "alpine", "postgres:15.17-alpine"},
	{"15", "debian", "postgres:15.17-bookworm"},
	{"16", "alpine", "postgres:16.13-alpine"},
	{"16", "debian", "postgres:16.13-bookworm"},
	{"17", "alpine", "postgres:17.9-alpine"},
	{"17", "debian", "postgres:17.9-bookworm"},
}

type cellResult struct {
	cell     matrixCell
	ok       bool
	duration time.Duration
	message  string
}

// Ping runs `SELECT version();` against the given DSN and returns the row as a string.
func (m *MatrixPgCompatGo) Ping(ctx context.Context, dsn string) (string, error) {
	return dag.Container().
		From("alpine:3.21").
		WithExec([]string{"apk", "add", "--no-cache", "postgresql17-client"}).
		WithEnvVariable("PGCONNECT_TIMEOUT", "30").
		WithExec([]string{"sh", "-c", fmt.Sprintf("psql %q -At -c 'SELECT version();'", dsn)}).
		Stdout(ctx)
}

// TestMatrix launches the 6-cell PG compatibility matrix in parallel.
// Returns a formatted ASCII table. Exits with an error if any cell fails.
func (m *MatrixPgCompatGo) TestMatrix(ctx context.Context) (string, error) {
	results := make([]cellResult, len(matrixCells))
	var wg sync.WaitGroup
	for i, cell := range matrixCells {
		wg.Add(1)
		go func(i int, cell matrixCell) {
			defer wg.Done()
			start := time.Now()
			output, err := m.runCell(ctx, cell)
			dur := time.Since(start)
			res := cellResult{cell: cell, duration: dur}
			if err != nil {
				res.message = err.Error()
			} else {
				needle := fmt.Sprintf("PostgreSQL %s.", cell.major)
				if strings.Contains(output, needle) {
					res.ok = true
					res.message = "matched " + needle
				} else {
					res.message = "expected " + needle + ", got: " + strings.TrimSpace(output)
				}
			}
			results[i] = res
		}(i, cell)
	}
	wg.Wait()

	table := formatResults(results)
	if countFailures(results) > 0 {
		return table, fmt.Errorf("%d cell(s) failed", countFailures(results))
	}
	return table, nil
}

func (m *MatrixPgCompatGo) runCell(ctx context.Context, cell matrixCell) (string, error) {
	pg := dag.Container().
		From(cell.tag).
		WithEnvVariable("POSTGRES_HOST_AUTH_METHOD", "trust").
		WithExposedPort(5432).
		AsService()

	client := dag.Container().
		From("alpine:3.21").
		WithExec([]string{"apk", "add", "--no-cache", "postgresql17-client"}).
		WithServiceBinding("pg", pg).
		WithExec([]string{"sh", "-c",
			"for i in $(seq 1 30); do pg_isready -h pg -U postgres && break; sleep 1; done; " +
				"psql 'postgres://postgres@pg:5432/postgres' -At -c 'SELECT version();'",
		})

	return client.Stdout(ctx)
}

func countFailures(results []cellResult) int {
	n := 0
	for _, r := range results {
		if !r.ok {
			n++
		}
	}
	return n
}

func formatResults(results []cellResult) string {
	var b strings.Builder
	b.WriteString("+----------+--------+--------+--------+\n")
	b.WriteString("| PG major | Base   | Status | Time   |\n")
	b.WriteString("+----------+--------+--------+--------+\n")
	ok := 0
	for _, r := range results {
		status := "FAIL"
		if r.ok {
			status = "OK"
			ok++
		}
		b.WriteString(fmt.Sprintf("| %-8s | %-6s | %-6s | %5.1fs |\n",
			r.cell.major, r.cell.base, status, r.duration.Seconds()))
	}
	b.WriteString("+----------+--------+--------+--------+\n")
	b.WriteString(fmt.Sprintf("%d/%d passed\n", ok, len(results)))
	for _, r := range results {
		if !r.ok {
			b.WriteString(fmt.Sprintf("FAIL %s/%s: %s\n", r.cell.major, r.cell.base, r.message))
		}
	}
	return b.String()
}
