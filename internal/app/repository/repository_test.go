package repository

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewRepository(t *testing.T) {
	ctx := context.Background()
	logger := zap.NewNop()

	t.Run("database DSN provided", func(t *testing.T) {
		repository, closer, err := NewRepository(ctx, logger, "", "invalid_dsn")
		assert.Error(t, err)
		assert.Nil(t, repository)
		assert.Nil(t, closer)
	})

	t.Run("file storage path provided", func(t *testing.T) {
		repository, closer, err := NewRepository(ctx, logger, "valid_path", "")
		assert.NoError(t, err)
		assert.NotNil(t, repository)
		assert.NotNil(t, closer)

		err = closer()
		assert.NoError(t, err)

		err = os.Remove("valid_path")
		assert.NoError(t, err)
	})

	t.Run("no database DSN or file storage path provided", func(t *testing.T) {
		repository, closer, err := NewRepository(ctx, logger, "", "")
		assert.NoError(t, err)
		assert.NotNil(t, repository)
		assert.Nil(t, closer)
	})
}
