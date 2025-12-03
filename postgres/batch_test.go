package postgres

import (
	"database/sql/driver"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
	"testing"
)

// BatchLoaderEntry for testing.
type TestBatchLoaderEntry struct {
	bl *BatchLoader[*TestBatchLoaderEntry]

	// Data to insert
	Data map[string]any
}

// Commit pushes the new entry into the batch loader.
func (e *TestBatchLoaderEntry) Commit() error {
	e.bl.Add(e.Data)
	return nil
}

func NewTestBatchLoaderEntry(bl *BatchLoader[*TestBatchLoaderEntry]) *TestBatchLoaderEntry {
	return &TestBatchLoaderEntry{
		bl:   bl,
		Data: make(map[string]any),
	}
}

type testProvider struct {
	db *sqlx.DB
}

func (tp *testProvider) GetDb() sqlx.Ext {
	return tp.db
}

func GenerateMockProvider(t *testing.T) (DBProvider, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	dbx := sqlx.NewDb(db, "sqlmock")

	return &testProvider{dbx}, mock, func() {
		db.Close()
	}
}

func TestBatchLoader(t *testing.T) {

	t.Parallel()

	t.Run("Standard batch loading", func(t *testing.T) {
		prov, mock, cleanup := GenerateMockProvider(t)
		defer cleanup()

		expectedArgs := make([]driver.Value, 0, 100)
		for i := 0; i < 100; i++ {
			expectedArgs = append(expectedArgs, i)
		}

		// We expect one query call with a single set of values.
		mock.ExpectExec("INSERT INTO test_table \\(amount\\) VALUES \\(").
			WithArgs(expectedArgs...).
			WillReturnResult(sqlmock.NewResult(0, 0))

		loader := NewBatchLoader(
			prov,
			NewTestBatchLoaderEntry,
			"test_table",
		)

		for i := 0; i < 100; i++ {
			item := loader.New()
			item.Data["amount"] = i
			require.NoError(t, item.Commit())
		}

		require.NoError(t, loader.Flush())
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Batch Limit", func(t *testing.T) {
		prov, mock, cleanup := GenerateMockProvider(t)
		defer cleanup()

		expectedArgs := make([]driver.Value, 0, 50)
		for i := 0; i < 50; i++ {
			expectedArgs = append(expectedArgs, i)
		}

		// We expect one query call with a single set of values.
		mock.ExpectExec("INSERT INTO test_table \\(amount\\) VALUES \\(").
			WithArgs(expectedArgs...).
			WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectExec("INSERT INTO test_table \\(amount\\) VALUES \\(").
			WithArgs(expectedArgs...).
			WillReturnResult(sqlmock.NewResult(0, 0))

		loader := NewBatchLoader(
			prov,
			NewTestBatchLoaderEntry,
			"test_table",
		).SetMaxValues(50)

		for i := 0; i < 100; i++ {
			item := loader.New()
			item.Data["amount"] = i % 50
			require.NoError(t, item.Commit())
		}

		require.NoError(t, loader.Flush())
		require.NoError(t, mock.ExpectationsWereMet())

	})

	t.Run("Defaults", func(t *testing.T) {
		prov, mock, cleanup := GenerateMockProvider(t)
		defer cleanup()

		expectedArgs := make([]driver.Value, 0, 100)
		for i := 0; i < 100; i++ {
			if i < 51 {
				expectedArgs = append(expectedArgs, i)
			} else {
				expectedArgs = append(expectedArgs, 999)
			}
		}

		// We expect one query call with a single set of values.
		mock.ExpectExec("INSERT INTO test_table \\(amount\\) VALUES \\(").
			WithArgs(expectedArgs...).
			WillReturnResult(sqlmock.NewResult(0, 0))

		loader := NewBatchLoader(
			prov,
			NewTestBatchLoaderEntry,
			"test_table",
		).SetDefaults(map[string]any{
			"amount": 999,
		})

		for i := 0; i < 100; i++ {
			item := loader.New()
			if i < 51 {
				item.Data["amount"] = i
			}
			require.NoError(t, item.Commit())
		}

		require.NoError(t, loader.Flush())
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Default to nil", func(t *testing.T) {
		prov, mock, cleanup := GenerateMockProvider(t)
		defer cleanup()

		expectedArgs := make([]driver.Value, 0, 100)
		for i := 0; i < 100; i++ {
			if i < 51 {
				expectedArgs = append(expectedArgs, i)
			} else {
				expectedArgs = append(expectedArgs, nil)
			}
		}

		// We expect one query call with a single set of values.
		mock.ExpectExec("INSERT INTO test_table \\(amount\\) VALUES \\(").
			WithArgs(expectedArgs...).
			WillReturnResult(sqlmock.NewResult(0, 0))

		loader := NewBatchLoader(
			prov,
			NewTestBatchLoaderEntry,
			"test_table",
		)

		for i := 0; i < 100; i++ {
			item := loader.New()
			if i < 51 {
				item.Data["amount"] = i
			}
			require.NoError(t, item.Commit())
		}

		require.NoError(t, loader.Flush())
		require.NoError(t, mock.ExpectationsWereMet())
	})
}
