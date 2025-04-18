package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
)

func TestNewPool(t *testing.T) {
	t.Run("error new pool", func(t *testing.T) {
		pool, err := NewPool(context.Background(), "")
		assert.NoError(t, err)
		assert.NotEmpty(t, pool)
	})
}

func TestRowScan(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()

	ctx := context.Background()
	sql := "SELECT 1"

	t.Run("successful scan", func(t *testing.T) {
		mockPool.ExpectQuery(sql).WillReturnRows(pgxmock.NewRows([]string{"column"}).AddRow(1))

		row := mockPool.QueryRow(ctx, sql)
		wrappedRow := &Row{
			Row: row,
			retry: func() pgx.Row {
				return mockPool.QueryRow(ctx, sql)
			},
		}

		var result int
		err := wrappedRow.Scan(&result)
		assert.NoError(t, err)
		assert.Equal(t, 1, result)
	})

	t.Run("retry on connection closed", func(t *testing.T) {
		mockPool.ExpectQuery(sql).WillReturnError(errors.New(connClosed))
		mockPool.ExpectQuery(sql).WillReturnRows(pgxmock.NewRows([]string{"column"}).AddRow(1))

		row := mockPool.QueryRow(ctx, sql)
		wrappedRow := &Row{
			Row: row,
			retry: func() pgx.Row {
				return mockPool.QueryRow(ctx, sql)
			},
		}

		var result int
		err := wrappedRow.Scan(&result)
		assert.NoError(t, err)
		assert.Equal(t, 1, result)
	})
}

func TestQueryRow(t *testing.T) {
	t.Run("error query row", func(t *testing.T) {
		ctx := context.Background()
		pgxpoolVar, err := pgxpool.New(ctx, "")
		assert.NoError(t, err)
		pool := Pool{Pool: pgxpoolVar}
		row := pool.QueryRow(ctx, "")
		assert.NotEmpty(t, row)
	})
}

func TestBegin(t *testing.T) {
	t.Run("error begin", func(t *testing.T) {
		ctx := context.Background()
		pgxpoolVar, err := pgxpool.New(ctx, "")
		assert.NoError(t, err)
		pool := Pool{Pool: pgxpoolVar}
		tx, err := pool.Begin(ctx)
		assert.Error(t, err)
		assert.Empty(t, tx)
	})
}

func TestTxQuery(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()

	ctx := context.Background()
	sql := "SELECT 1"

	t.Run("successful query", func(t *testing.T) {
		mockPool.ExpectBegin()
		mockPool.ExpectQuery(sql).WillReturnRows(pgxmock.NewRows([]string{"column"}).AddRow(1))

		tx, err := mockPool.Begin(ctx)
		assert.NoError(t, err)

		wrappedTx := &Tx{
			Tx: tx,
		}

		rows, err := wrappedTx.Query(ctx, sql)
		assert.NoError(t, err)
		assert.NotNil(t, rows)

		rows.Close()
	})

	t.Run("retry on connection closed", func(t *testing.T) {
		mockPool.ExpectBegin()
		mockPool.ExpectQuery(sql).WillReturnError(errors.New(connClosed))
		mockPool.ExpectQuery(sql).WillReturnRows(pgxmock.NewRows([]string{"column"}).AddRow(1))

		tx, err := mockPool.Begin(ctx)
		assert.NoError(t, err)

		wrappedTx := &Tx{
			Tx: tx,
		}

		rows, err := wrappedTx.Query(ctx, sql)
		assert.NoError(t, err)
		assert.NotNil(t, rows)

		rows.Close()
	})

	t.Run("retry exhausted", func(t *testing.T) {
		mockPool.ExpectBegin()
		mockPool.ExpectQuery(sql).WillReturnError(errors.New(connClosed))

		tx, err := mockPool.Begin(ctx)
		assert.NoError(t, err)

		wrappedTx := &Tx{
			Tx: tx,
		}

		rows, err := wrappedTx.Query(ctx, sql)
		assert.Error(t, err)
		assert.Nil(t, rows)

		if rows != nil {
			rows.Close()
		}
	})
}

func TestBatchResultsExec(t *testing.T) {
	mockPool, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mockPool.Close()

	ctx := context.Background()
	batch := &pgx.Batch{
		QueuedQueries: []*pgx.QueuedQuery{{SQL: "INSERT INTO test"}},
	}

	t.Run("successful exec", func(t *testing.T) {
		mockPool.ExpectBegin()
		mockPool.ExpectBatch().ExpectExec("INSERT INTO test").WillReturnResult(pgxmock.NewResult("INSERT", 1))
		mockPool.ExpectCommit()

		tx, err := mockPool.Begin(ctx)
		assert.NoError(t, err)

		wrappedTx := &Tx{
			Tx: tx,
		}

		batchResults := wrappedTx.SendBatch(ctx, batch)
		wrappedBatchResults := &BatchResults{
			BatchResults: batchResults,
			retry: func() pgx.BatchResults {
				return tx.SendBatch(ctx, batch)
			},
		}

		commandTag, err := wrappedBatchResults.Exec()
		assert.NoError(t, err)
		assert.Equal(t, "INSERT 1", commandTag.String())

		err = wrappedTx.Commit(ctx)
		assert.NoError(t, err)
	})
}

func TestPing(t *testing.T) {
	t.Run("error ping", func(t *testing.T) {
		ctx := context.Background()
		pgxpoolVar, err := pgxpool.New(ctx, "")
		assert.NoError(t, err)
		pool := Pool{Pool: pgxpoolVar}
		err = pool.Ping(ctx)
		assert.Error(t, err)
	})
}
