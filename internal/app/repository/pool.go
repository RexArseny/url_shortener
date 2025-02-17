//nolint:wrapcheck // methods have been overridden so errors passed through
package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Retry variables.
const (
	retry      = 3
	connClosed = "conn closed"
)

// Pool is a wrapper for database Pool.
type Pool struct {
	*pgxpool.Pool
}

// Row is a wrapper for database Row.
type Row struct {
	pgx.Row
	retry func() pgx.Row
}

// Tx is a wrapper for database Tx.
type Tx struct {
	pgx.Tx
}

// BatchResults is a wrapper for database BatchResults.
type BatchResults struct {
	pgx.BatchResults
	retry func() pgx.BatchResults
}

// NewPool create new pool.
func NewPool(ctx context.Context, connString string) (*Pool, error) {
	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, fmt.Errorf("can not create new pool for PostgreSQL server: %w", err)
	}
	return &Pool{
		Pool: pool,
	}, nil
}

// Scan is a wrapper for database Scan.
func (r *Row) Scan(dest ...any) error {
	var err error
	for range retry {
		err = r.Row.Scan(dest...)
		if err == nil {
			return nil
		}
		if !strings.Contains(err.Error(), connClosed) {
			return err
		}
		r.Row = r.retry()
	}

	return err
}

// QueryRow is a wrapper for database QueryRow.
func (p *Pool) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	row := p.Pool.QueryRow(ctx, sql, args...)
	return &Row{
		Row: row,
		retry: func() pgx.Row {
			return p.Pool.QueryRow(ctx, sql, args...)
		},
	}
}

// Begin is a wrapper for database Begin.
func (p *Pool) Begin(ctx context.Context) (pgx.Tx, error) {
	var err error
	for range retry {
		var tx pgx.Tx
		tx, err = p.Pool.Begin(ctx)
		if err == nil {
			return &Tx{
				Tx: tx,
			}, nil
		}
		if !strings.Contains(err.Error(), connClosed) {
			return nil, err
		}
	}

	return nil, err
}

// Query is a wrapper for database Query.
func (t *Tx) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	var err error
	for range retry {
		var rows pgx.Rows
		rows, err = t.Tx.Query(ctx, sql, args...)
		if err == nil {
			return rows, nil
		}
		if !strings.Contains(err.Error(), connClosed) {
			return nil, err
		}
	}

	return nil, err
}

// SendBatch is a wrapper for database SendBatch.
func (t *Tx) SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults {
	batchResults := t.Tx.SendBatch(ctx, b)
	return &BatchResults{
		BatchResults: batchResults,
		retry: func() pgx.BatchResults {
			return t.Tx.SendBatch(ctx, b)
		},
	}
}

// Commit is a wrapper for database Commit.
func (t *Tx) Commit(ctx context.Context) error {
	var err error
	for range retry {
		err = t.Tx.Commit(ctx)
		if err == nil {
			return nil
		}
		if !strings.Contains(err.Error(), connClosed) {
			return err
		}
	}

	return err
}

// Exec is a wrapper for database Exec.
func (b *BatchResults) Exec() (pgconn.CommandTag, error) {
	var err error
	for range retry {
		var commandTag pgconn.CommandTag
		commandTag, err = b.BatchResults.Exec()
		if err == nil {
			return commandTag, nil
		}
		if !strings.Contains(err.Error(), connClosed) {
			return pgconn.CommandTag{}, err
		}
		b.BatchResults = b.retry()
	}

	return pgconn.CommandTag{}, err
}

// Ping is a wrapper for database Ping.
func (p *Pool) Ping(ctx context.Context) error {
	var err error
	for range retry {
		err = p.Pool.Ping(ctx)
		if err == nil {
			return nil
		}
		if !strings.Contains(err.Error(), connClosed) {
			return err
		}
	}

	return err
}
