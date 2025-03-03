package repository

import (
	"context"
	"reflect"
	"testing"

	"github.com/RexArseny/url_shortener/internal/app/logger"
	"github.com/RexArseny/url_shortener/internal/app/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// MockPool is a mock of Pool interface.
type MockPool struct {
	ctrl     *gomock.Controller
	recorder *MockPoolMockRecorder
}

// MockPoolMockRecorder is the mock recorder for MockPool.
type MockPoolMockRecorder struct {
	mock *MockPool
}

// NewMockPool creates a new mock instance.
func NewMockPool(ctrl *gomock.Controller) *MockPool {
	mock := &MockPool{ctrl: ctrl}
	mock.recorder = &MockPoolMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockPool) EXPECT() *MockPoolMockRecorder {
	return m.recorder
}

// QueryRow mocks the QueryRow method.
func (m *MockPool) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, sql}
	varargs = append(varargs, args...)
	ret := m.ctrl.Call(m, "QueryRow", varargs...)
	ret0, _ := ret[0].(pgx.Row)
	return ret0
}

// QueryRow indicates an expected call of QueryRow.
func (mr *MockPoolMockRecorder) QueryRow(ctx interface{}, sql interface{}, args ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, sql}, args...)
	return mr.mock.ctrl.RecordCallWithMethodType(
		mr.mock,
		"QueryRow",
		reflect.TypeOf((*MockPool)(nil).QueryRow),
		varargs...)
}

// Query mocks the Query method.
func (m *MockPool) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, sql}
	varargs = append(varargs, args...)
	ret := m.ctrl.Call(m, "Query", varargs...)
	ret0, _ := ret[0].(pgx.Rows)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Query indicates an expected call of Query.
func (mr *MockPoolMockRecorder) Query(ctx interface{}, sql interface{}, args ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, sql}, args...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Query", reflect.TypeOf((*MockPool)(nil).Query), varargs...)
}

// Exec mocks the Exec method.
func (m *MockPool) Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, sql}
	varargs = append(varargs, args...)
	ret := m.ctrl.Call(m, "Exec", varargs...)
	ret0, _ := ret[0].(pgconn.CommandTag)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Exec indicates an expected call of Exec.
func (mr *MockPoolMockRecorder) Exec(ctx interface{}, sql interface{}, args ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, sql}, args...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Exec", reflect.TypeOf((*MockPool)(nil).Exec), varargs...)
}

// Begin mocks the Begin method.
func (m *MockPool) Begin(ctx context.Context) (pgx.Tx, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Begin", ctx)
	ret0, _ := ret[0].(pgx.Tx)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Begin indicates an expected call of Begin.
func (mr *MockPoolMockRecorder) Begin(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Begin", reflect.TypeOf((*MockPool)(nil).Begin), ctx)
}

// Ping mocks the Ping method.
func (m *MockPool) Ping(ctx context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Ping", ctx)
	ret0, _ := ret[0].(error)
	return ret0
}

// Ping indicates an expected call of Ping.
func (mr *MockPoolMockRecorder) Ping(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Ping", reflect.TypeOf((*MockPool)(nil).Ping), ctx)
}

// Close mocks the Close method.
func (m *MockPool) Close() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Close")
}

// Close indicates an expected call of Close.
func (mr *MockPoolMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockPool)(nil).Close))
}

func TestDBRepositoryGetOriginalURL(t *testing.T) {
	testLogger, err := logger.InitLogger()
	assert.NoError(t, err)

	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	repo := &DBRepository{
		logger: testLogger.Named("repository"),
		pool:   mock,
	}

	shortLink := "abc123"
	originalURL := "http://example.com"
	deleted := false

	mock.ExpectQuery("SELECT original_url, deleted FROM urls WHERE short_url=").
		WithArgs(shortLink).
		WillReturnRows(pgxmock.NewRows([]string{"original_url", "deleted"}).
			AddRow(originalURL, deleted))

	result, err := repo.GetOriginalURL(context.Background(), shortLink)
	assert.NoError(t, err)
	assert.Equal(t, originalURL, *result)
}

func TestDBRepositorySetLink(t *testing.T) {
	testLogger, err := logger.InitLogger()
	assert.NoError(t, err)

	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	repo := &DBRepository{
		logger: testLogger.Named("repository"),
		pool:   mock,
	}

	originalURL := "http://example.com"
	shortURL := "abc123"
	userID := uuid.New()

	mock.ExpectQuery("INSERT INTO urls").
		WithArgs(shortURL, originalURL, userID).
		WillReturnRows(pgxmock.NewRows([]string{"short_url"}).AddRow(shortURL))

	result, err := repo.SetLink(context.Background(), originalURL, []string{shortURL}, userID)
	assert.NoError(t, err)
	assert.Equal(t, shortURL, *result)
}

func TestDBRepositorySetLinks(t *testing.T) {
	testLogger, err := logger.InitLogger()
	assert.NoError(t, err)

	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	repo := &DBRepository{
		logger: testLogger.Named("repository"),
		pool:   mock,
	}

	batch := []models.ShortenBatchRequest{
		{OriginalURL: "http://example.com"},
	}
	shortURLs := [][]string{{"abc123"}}
	userID := uuid.New()

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT short_url, original_url FROM urls WHERE original_url = ANY").
		WithArgs(pgxmock.AnyArg()).
		WillReturnRows(pgxmock.NewRows([]string{"short_url", "original_url"}))
	mock.ExpectBatch().ExpectExec("INSERT INTO urls").
		WithArgs(shortURLs[0][0], batch[0].OriginalURL, userID).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))
	mock.ExpectCommit()

	result, err := repo.SetLinks(context.Background(), batch, shortURLs, userID)
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestDBRepositoryGetShortLinksOfUser(t *testing.T) {
	testLogger, err := logger.InitLogger()
	assert.NoError(t, err)

	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	repo := &DBRepository{
		logger: testLogger.Named("repository"),
		pool:   mock,
	}

	userID := uuid.New()

	mock.ExpectQuery("SELECT short_url, original_url FROM urls WHERE user_id =").
		WithArgs(userID).
		WillReturnRows(pgxmock.NewRows([]string{"short_url", "original_url"}).
			AddRow("abc123", "http://example.com"))

	result, err := repo.GetShortLinksOfUser(context.Background(), userID)
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestDBRepositoryDeleteURLs(t *testing.T) {
	testLogger, err := logger.InitLogger()
	assert.NoError(t, err)

	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	repo := &DBRepository{
		logger: testLogger.Named("repository"),
		pool:   mock,
	}

	urls := []string{"abc123"}
	userID := uuid.New()

	mock.ExpectExec("INSERT INTO urls_for_delete").
		WithArgs(urls, userID).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	err = repo.DeleteURLs(context.Background(), urls, userID)
	assert.NoError(t, err)
}

func TestDBRepositoryDeleteURLsInDB(t *testing.T) {
	testLogger, err := logger.InitLogger()
	assert.NoError(t, err)

	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	repo := &DBRepository{
		logger: testLogger.Named("repository"),
		pool:   mock,
	}

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id, urls, user_id FROM urls_for_delete LIMIT 1").
		WillReturnRows(pgxmock.NewRows([]string{"id", "urls", "user_id"}).
			AddRow(1, []string{"abc123"}, uuid.New()))
	mock.ExpectExec("UPDATE urls SET deleted = true").
		WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	mock.ExpectExec("DELETE FROM urls_for_delete").
		WithArgs(1).
		WillReturnResult(pgxmock.NewResult("DELETE", 1))
	mock.ExpectCommit()

	err = repo.DeleteURLsInDB(context.Background())
	assert.NoError(t, err)
}

func TestDBRepositoryPing(t *testing.T) {
	testLogger, err := logger.InitLogger()
	assert.NoError(t, err)

	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	repo := &DBRepository{
		logger: testLogger.Named("repository"),
		pool:   mock,
	}

	mock.ExpectPing()

	err = repo.Ping(context.Background())
	assert.NoError(t, err)
}

func TestDBRepositoryClose(t *testing.T) {
	testLogger, err := logger.InitLogger()
	assert.NoError(t, err)

	mock, err := pgxmock.NewPool()
	assert.NoError(t, err)
	defer mock.Close()

	repo := &DBRepository{
		logger: testLogger.Named("repository"),
		pool:   mock,
	}

	mock.ExpectClose()

	repo.Close()
}
