package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
)

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

		tx, err := mockPool.Begin(ctx)
		assert.NoError(t, err)

		batchResults := tx.SendBatch(ctx, batch)
		wrappedBatchResults := &BatchResults{
			BatchResults: batchResults,
			retry: func() pgx.BatchResults {
				return tx.SendBatch(ctx, batch)
			},
		}

		commandTag, err := wrappedBatchResults.Exec()
		assert.NoError(t, err)
		assert.Equal(t, "INSERT 1", commandTag.String())
	})
}
