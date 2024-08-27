package sqlstorage_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/numary/ledger/internal/pgtesting"
	"github.com/numary/ledger/pkg/api/idempotency"
	"github.com/numary/ledger/pkg/core"
	"github.com/numary/ledger/pkg/ledger"
	"github.com/numary/ledger/pkg/ledgertesting"
	"github.com/numary/ledger/pkg/storage"
	"github.com/numary/ledger/pkg/storage/sqlstorage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
)

func TestStore(t *testing.T) {
	type testingFunction struct {
		name string
		fn   func(t *testing.T, store *sqlstorage.Store)
	}

	runTest := func(tf testingFunction) func(t *testing.T) {
		return func(t *testing.T) {
			done := make(chan struct{})
			app := fx.New(
				ledgertesting.ProvideStorageDriver(false),
				fx.NopLogger,
				fx.Invoke(func(driver *sqlstorage.Driver, lc fx.Lifecycle) {
					lc.Append(fx.Hook{
						OnStart: func(ctx context.Context) error {
							defer func() {
								close(done)
							}()
							store, _, err := driver.GetLedgerStore(ctx, uuid.NewString(), true)
							if err != nil {
								return err
							}
							defer store.Close(ctx)

							if _, err = store.Initialize(context.Background()); err != nil {
								return err
							}

							tf.fn(t, store)
							return nil
						},
					})
				}),
			)
			go func() {
				require.NoError(t, app.Start(context.Background()))
			}()
			defer func(app *fx.App, ctx context.Context) {
				require.NoError(t, app.Stop(ctx))
			}(app, context.Background())

			select {
			case <-time.After(5 * time.Second):
				t.Fatal("timeout")
			case <-done:
			}
		}
	}

	for _, tf := range []testingFunction{
		{name: "Accounts", fn: testAccounts},
		{name: "Commit", fn: testCommit},
		{name: "UpdateTransactionMetadata", fn: testUpdateTransactionMetadata},
		{name: "UpdateAccountMetadata", fn: testUpdateAccountMetadata},
		{name: "GetLastLog", fn: testGetLastLog},
		{name: "GetLogs", fn: testGetLogs},
		{name: "CountAccounts", fn: testCountAccounts},
		{name: "GetAssetsVolumes", fn: testGetAssetsVolumes},
		{name: "GetAccounts", fn: testGetAccounts},
		{name: "Transactions", fn: testTransactions},
		{name: "GetTransaction", fn: testGetTransaction},
		{name: "GetTransactionWithQueryAddress", fn: testTransactionsQueryAddress},
		{name: "Mapping", fn: testMapping},
		{name: "TooManyClient", fn: testTooManyClient},
		{name: "GetBalances", fn: testGetBalances},
		{name: "GetBalances1Accounts", fn: testGetBalancesOn1Account},
		{name: "GetBalancesBigInts", fn: testGetBalancesBigInts},
		{name: "GetBalancesAggregated", fn: testGetBalancesAggregated},
		{name: "GetBalancesAggregatedByAccount", fn: testGetBalancesAggregatedByAccount},
		{name: "CreateIK", fn: testIKS},
		{name: "GetTransactionsByAccount", fn: testGetTransactionsByAccount},
	} {
		t.Run(fmt.Sprintf("%s/%s-singleInstance", ledgertesting.StorageDriverName(), tf.name), runTest((tf)))
	}
}

var now = time.Now().UTC().Truncate(time.Second)
var tx1 = core.ExpandedTransaction{
	Transaction: core.Transaction{
		TransactionData: core.TransactionData{
			Postings: []core.Posting{
				{
					Source:      "world",
					Destination: "central_bank",
					Amount:      core.NewMonetaryInt(100),
					Asset:       "USD",
				},
			},
			Reference: "tx1",
			Timestamp: now.Add(-3 * time.Hour),
		},
	},
	PostCommitVolumes: core.AccountsAssetsVolumes{
		"world": {
			"USD": {
				Input:  core.NewMonetaryInt(0),
				Output: core.NewMonetaryInt(100),
			},
		},
		"central_bank": {
			"USD": {
				Input:  core.NewMonetaryInt(100),
				Output: core.NewMonetaryInt(0),
			},
		},
	},
	PreCommitVolumes: core.AccountsAssetsVolumes{
		"world": {
			"USD": {
				Input:  core.NewMonetaryInt(0),
				Output: core.NewMonetaryInt(0),
			},
		},
		"central_bank": {
			"USD": {
				Input:  core.NewMonetaryInt(0),
				Output: core.NewMonetaryInt(0),
			},
		},
	},
}
var tx2 = core.ExpandedTransaction{
	Transaction: core.Transaction{
		ID: 1,
		TransactionData: core.TransactionData{
			Postings: []core.Posting{
				{
					Source:      "world",
					Destination: "central_bank",
					Amount:      core.NewMonetaryInt(100),
					Asset:       "USD",
				},
			},
			Reference: "tx2",
			Timestamp: now.Add(-2 * time.Hour),
		},
	},
	PostCommitVolumes: core.AccountsAssetsVolumes{
		"world": {
			"USD": {
				Input:  core.NewMonetaryInt(0),
				Output: core.NewMonetaryInt(200),
			},
		},
		"central_bank": {
			"USD": {
				Input:  core.NewMonetaryInt(200),
				Output: core.NewMonetaryInt(0),
			},
		},
	},
	PreCommitVolumes: core.AccountsAssetsVolumes{
		"world": {
			"USD": {
				Input:  core.NewMonetaryInt(0),
				Output: core.NewMonetaryInt(100),
			},
		},
		"central_bank": {
			"USD": {
				Input:  core.NewMonetaryInt(100),
				Output: core.NewMonetaryInt(0),
			},
		},
	},
}
var tx3 = core.ExpandedTransaction{
	Transaction: core.Transaction{
		ID: 2,
		TransactionData: core.TransactionData{
			Postings: []core.Posting{
				{
					Source:      "central_bank",
					Destination: "users:1",
					Amount:      core.NewMonetaryInt(1),
					Asset:       "USD",
				},
			},
			Reference: "tx3",
			Metadata: core.Metadata{
				"priority": json.RawMessage(`"high"`),
			},
			Timestamp: now.Add(-1 * time.Hour),
		},
	},
	PreCommitVolumes: core.AccountsAssetsVolumes{
		"central_bank": {
			"USD": {
				Input:  core.NewMonetaryInt(200),
				Output: core.NewMonetaryInt(0),
			},
		},
		"users:1": {
			"USD": {
				Input:  core.NewMonetaryInt(0),
				Output: core.NewMonetaryInt(0),
			},
		},
	},
	PostCommitVolumes: core.AccountsAssetsVolumes{
		"central_bank": {
			"USD": {
				Input:  core.NewMonetaryInt(200),
				Output: core.NewMonetaryInt(1),
			},
		},
		"users:1": {
			"USD": {
				Input:  core.NewMonetaryInt(1),
				Output: core.NewMonetaryInt(0),
			},
		},
	},
}

var tx4 = core.ExpandedTransaction{
	Transaction: core.Transaction{
		ID: 3,
		TransactionData: core.TransactionData{
			Postings: []core.Posting{
				{
					Source:      "central_bank",
					Destination: "users:11",
					Amount:      core.NewMonetaryInt(1),
					Asset:       "USD",
				},
			},
			Reference: "tx4",
			Metadata: core.Metadata{
				"priority": json.RawMessage(`"high"`),
			},
			Timestamp: now.Add(-1 * time.Hour),
		},
	},
	PreCommitVolumes: core.AccountsAssetsVolumes{
		"central_bank": {
			"USD": {
				Input:  core.NewMonetaryInt(200),
				Output: core.NewMonetaryInt(0),
			},
		},
		"users:11": {
			"USD": {
				Input:  core.NewMonetaryInt(0),
				Output: core.NewMonetaryInt(0),
			},
		},
	},
	PostCommitVolumes: core.AccountsAssetsVolumes{
		"central_bank": {
			"USD": {
				Input:  core.NewMonetaryInt(200),
				Output: core.NewMonetaryInt(1),
			},
		},
		"users:11": {
			"USD": {
				Input:  core.NewMonetaryInt(1),
				Output: core.NewMonetaryInt(0),
			},
		},
	},
}

var tx5 = core.ExpandedTransaction{
	Transaction: core.Transaction{
		ID: 4,
		TransactionData: core.TransactionData{
			Postings: []core.Posting{
				{
					Source:      "users:1",
					Destination: "central_bank",
					Amount:      core.NewMonetaryInt(1),
					Asset:       "USD",
				},
			},
			Reference: "tx5",
			Metadata: core.Metadata{
				"priority": json.RawMessage(`"high"`),
			},
			Timestamp: now.Add(-1 * time.Hour),
		},
	},
	PreCommitVolumes: core.AccountsAssetsVolumes{
		"users:1": {
			"USD": {
				Input:  core.NewMonetaryInt(200),
				Output: core.NewMonetaryInt(0),
			},
		},
		"central_bank": {
			"USD": {
				Input:  core.NewMonetaryInt(0),
				Output: core.NewMonetaryInt(0),
			},
		},
	},
	PostCommitVolumes: core.AccountsAssetsVolumes{
		"users:!": {
			"USD": {
				Input:  core.NewMonetaryInt(200),
				Output: core.NewMonetaryInt(1),
			},
		},
		"central_bank": {
			"USD": {
				Input:  core.NewMonetaryInt(1),
				Output: core.NewMonetaryInt(0),
			},
		},
	},
}

func testCommit(t *testing.T, store *sqlstorage.Store) {
	tx := core.ExpandedTransaction{
		Transaction: core.Transaction{
			ID: 0,
			TransactionData: core.TransactionData{
				Postings: []core.Posting{
					{
						Source:      "world",
						Destination: "central_bank",
						Amount:      core.NewMonetaryInt(100),
						Asset:       "USD",
					},
				},
				Reference: "foo",
				Timestamp: time.Now().Round(time.Second),
			},
		},
	}
	err := store.Commit(context.Background(), tx)
	require.NoError(t, err)

	err = store.Commit(context.Background(), tx)
	require.Error(t, err)
	require.True(t, storage.IsErrorCode(err, storage.ConstraintTXID))

	cursor, err := store.GetLogs(context.Background(), ledger.NewLogsQuery())
	require.NoError(t, err)
	require.Len(t, cursor.Data, 1)
}

func testIKS(t *testing.T, store *sqlstorage.Store) {
	t.Run("Create and Read", func(t *testing.T) {
		response := idempotency.Response{
			RequestHash: "xxx",
			StatusCode:  http.StatusAccepted,
			Header: http.Header{
				"Content-Type": []string{"application/json"},
			},
			Body: "Hello World!",
		}
		require.NoError(t, store.CreateIK(context.Background(), "foo", response))

		fromDB, err := store.ReadIK(context.Background(), "foo")
		require.NoError(t, err)
		require.Equal(t, response, *fromDB)
	})
	t.Run("Not found", func(t *testing.T) {
		_, err := store.ReadIK(context.Background(), uuid.NewString())
		require.Equal(t, idempotency.ErrIKNotFound, err)
	})
}

func testUpdateTransactionMetadata(t *testing.T, store *sqlstorage.Store) {
	tx := core.ExpandedTransaction{
		Transaction: core.Transaction{
			ID: 0,
			TransactionData: core.TransactionData{
				Postings: []core.Posting{
					{
						Source:      "world",
						Destination: "central_bank",
						Amount:      core.NewMonetaryInt(100),
						Asset:       "USD",
					},
				},
				Reference: "foo",
				Timestamp: time.Now().Round(time.Second),
			},
		},
	}
	err := store.Commit(context.Background(), tx)
	require.NoError(t, err)

	err = store.UpdateTransactionMetadata(context.Background(), tx.ID, core.Metadata{
		"foo": "bar",
	}, time.Now())
	require.NoError(t, err)

	retrievedTransaction, err := store.GetTransaction(context.Background(), tx.ID)
	require.NoError(t, err)
	require.EqualValues(t, "bar", retrievedTransaction.Metadata["foo"])

	cursor, err := store.GetLogs(context.Background(), ledger.NewLogsQuery())
	require.NoError(t, err)
	require.Len(t, cursor.Data, 2)
}

func testUpdateAccountMetadata(t *testing.T, store *sqlstorage.Store) {
	tx := core.ExpandedTransaction{
		Transaction: core.Transaction{
			ID: 0,
			TransactionData: core.TransactionData{
				Postings: []core.Posting{
					{
						Source:      "world",
						Destination: "central_bank",
						Amount:      core.NewMonetaryInt(100),
						Asset:       "USD",
					},
				},
				Reference: "foo",
				Timestamp: time.Now().Round(time.Second),
			},
		},
	}
	err := store.Commit(context.Background(), tx)
	require.NoError(t, err)

	err = store.UpdateAccountMetadata(context.Background(), "central_bank", core.Metadata{
		"foo": "bar",
	}, time.Now())
	require.NoError(t, err)

	account, err := store.GetAccount(context.Background(), "central_bank")
	require.NoError(t, err)
	require.EqualValues(t, "bar", account.Metadata["foo"])

	cursor, err := store.GetLogs(context.Background(), ledger.NewLogsQuery())
	require.NoError(t, err)
	require.Len(t, cursor.Data, 2)
}

func testCountAccounts(t *testing.T, store *sqlstorage.Store) {
	tx := core.ExpandedTransaction{
		Transaction: core.Transaction{
			ID: 0,
			TransactionData: core.TransactionData{
				Postings: []core.Posting{
					{
						Source:      "world",
						Destination: "central_bank",
						Amount:      core.NewMonetaryInt(100),
						Asset:       "USD",
					},
				},
				Timestamp: time.Now().Round(time.Second),
			},
		},
	}
	err := store.Commit(context.Background(), tx)
	require.NoError(t, err)

	countAccounts, err := store.CountAccounts(context.Background(), ledger.AccountsQuery{})
	require.NoError(t, err)
	require.EqualValues(t, 2, countAccounts) // world + central_bank
}

func testGetAssetsVolumes(t *testing.T, store *sqlstorage.Store) {
	tx := core.ExpandedTransaction{
		Transaction: core.Transaction{
			TransactionData: core.TransactionData{
				Postings: []core.Posting{
					{
						Source:      "world",
						Destination: "central_bank",
						Amount:      core.NewMonetaryInt(100),
						Asset:       "USD",
					},
				},
				Timestamp: time.Now().Round(time.Second),
			},
		},
		PostCommitVolumes: core.AccountsAssetsVolumes{
			"central_bank": core.AssetsVolumes{
				"USD": {
					Input:  core.NewMonetaryInt(100),
					Output: core.NewMonetaryInt(0),
				},
			},
		},
		PreCommitVolumes: core.AccountsAssetsVolumes{
			"central_bank": core.AssetsVolumes{
				"USD": {
					Input:  core.NewMonetaryInt(100),
					Output: core.NewMonetaryInt(0),
				},
			},
		},
	}
	err := store.Commit(context.Background(), tx)
	require.NoError(t, err)

	volumes, err := store.GetAssetsVolumes(context.Background(), "central_bank")
	require.NoError(t, err)
	require.Len(t, volumes, 1)
	require.EqualValues(t, core.NewMonetaryInt(100), volumes["USD"].Input)
	require.EqualValues(t, core.NewMonetaryInt(0), volumes["USD"].Output)
}

func testGetAccounts(t *testing.T, store *sqlstorage.Store) {
	require.NoError(t, store.UpdateAccountMetadata(context.Background(), "world", core.Metadata{
		"foo": json.RawMessage(`"bar"`),
	}, now))
	require.NoError(t, store.UpdateAccountMetadata(context.Background(), "bank", core.Metadata{
		"hello": json.RawMessage(`"world"`),
	}, now))
	require.NoError(t, store.UpdateAccountMetadata(context.Background(), "order:1", core.Metadata{
		"hello": json.RawMessage(`"world"`),
	}, now))
	require.NoError(t, store.UpdateAccountMetadata(context.Background(), "order:2", core.Metadata{
		"number":  json.RawMessage(`3`),
		"boolean": json.RawMessage(`true`),
		"a":       json.RawMessage(`{"super": {"nested": {"key": "hello"}}}`),
	}, now))

	accounts, err := store.GetAccounts(context.Background(), ledger.AccountsQuery{
		PageSize: 1,
	})
	require.NoError(t, err)
	require.Equal(t, 1, accounts.PageSize)
	require.Len(t, accounts.Data, 1)

	accounts, err = store.GetAccounts(context.Background(), ledger.AccountsQuery{
		PageSize:     1,
		AfterAddress: string(accounts.Data[0].Address),
	})
	require.NoError(t, err)
	require.Equal(t, 1, accounts.PageSize)

	accounts, err = store.GetAccounts(context.Background(), ledger.AccountsQuery{
		PageSize: 10,
		Filters: ledger.AccountsQueryFilters{
			Address: `^.*der:.*$`,
		},
	})
	require.NoError(t, err)
	require.Len(t, accounts.Data, 2)
	require.Equal(t, 10, accounts.PageSize)

	accounts, err = store.GetAccounts(context.Background(), ledger.AccountsQuery{
		PageSize: 10,
		Filters: ledger.AccountsQueryFilters{
			Metadata: map[string]string{
				"foo": "bar",
			},
		},
	})
	require.NoError(t, err)
	require.Len(t, accounts.Data, 1)

	accounts, err = store.GetAccounts(context.Background(), ledger.AccountsQuery{
		PageSize: 10,
		Filters: ledger.AccountsQueryFilters{
			Metadata: map[string]string{
				"number": "3",
			},
		},
	})
	require.NoError(t, err)
	require.Len(t, accounts.Data, 1)

	accounts, err = store.GetAccounts(context.Background(), ledger.AccountsQuery{
		PageSize: 10,
		Filters: ledger.AccountsQueryFilters{
			Metadata: map[string]string{
				"boolean": "true",
			},
		},
	})
	require.NoError(t, err)
	require.Len(t, accounts.Data, 1)

	accounts, err = store.GetAccounts(context.Background(), ledger.AccountsQuery{
		PageSize: 10,
		Filters: ledger.AccountsQueryFilters{
			Metadata: map[string]string{
				"a.super.nested.key": "hello",
			},
		},
	})
	require.NoError(t, err)
	require.Len(t, accounts.Data, 1)
}

func testMapping(t *testing.T, store *sqlstorage.Store) {
	m := core.Mapping{
		Contracts: []core.Contract{
			{
				Expr: &core.ExprGt{
					Op1: core.VariableExpr{Name: "balance"},
					Op2: core.ConstantExpr{Value: core.NewMonetaryInt(0)},
				},
				Account: "orders:*",
			},
		},
	}
	err := store.SaveMapping(context.Background(), m)
	assert.NoError(t, err)

	mapping, err := store.LoadMapping(context.Background())
	assert.NoError(t, err)
	assert.Len(t, mapping.Contracts, 1)
	assert.EqualValues(t, m.Contracts[0], mapping.Contracts[0])

	m2 := core.Mapping{
		Contracts: []core.Contract{},
	}
	err = store.SaveMapping(context.Background(), m2)
	assert.NoError(t, err)

	mapping, err = store.LoadMapping(context.Background())
	assert.NoError(t, err)
	assert.Len(t, mapping.Contracts, 0)
}

func testTooManyClient(t *testing.T, store *sqlstorage.Store) {
	// Use of external server, ignore this test
	if os.Getenv("NUMARY_STORAGE_POSTGRES_CONN_STRING") != "" ||
		ledgertesting.StorageDriverName() != "postgres" {
		return
	}

	for i := 0; i < pgtesting.MaxConnections; i++ {
		tx, err := store.Schema().BeginTx(context.Background(), nil)
		require.NoError(t, err)
		defer func(tx *sql.Tx) {
			require.NoError(t, tx.Rollback())
		}(tx)
	}

	_, err := store.CountTransactions(context.Background(), ledger.TransactionsQuery{})
	require.Error(t, err)
	require.IsType(t, new(storage.Error), err)
	require.Equal(t, storage.TooManyClient, err.(*storage.Error).Code)
}

func TestInitializeStore(t *testing.T) {
	driver, stopFn, err := ledgertesting.StorageDriver(false)
	require.NoError(t, err)
	defer stopFn()
	defer func(driver storage.Driver[*sqlstorage.Store], ctx context.Context) {
		require.NoError(t, driver.Close(ctx))
	}(driver, context.Background())

	err = driver.Initialize(context.Background())
	require.NoError(t, err)

	store, _, err := driver.GetLedgerStore(context.Background(), uuid.NewString(), true)
	require.NoError(t, err)

	modified, err := store.Initialize(context.Background())
	require.NoError(t, err)
	require.True(t, modified)

	modified, err = store.Initialize(context.Background())
	require.NoError(t, err)
	require.False(t, modified)
}

func testGetLastLog(t *testing.T, store *sqlstorage.Store) {
	err := store.Commit(context.Background(), tx1)
	require.NoError(t, err)

	lastLog, err := store.GetLastLog(context.Background())
	require.NoError(t, err)
	require.NotNil(t, lastLog)

	require.Equal(t, tx1.Postings, lastLog.Data.(core.Transaction).Postings)
	require.Equal(t, tx1.Reference, lastLog.Data.(core.Transaction).Reference)
	require.Equal(t, tx1.Timestamp, lastLog.Data.(core.Transaction).Timestamp)
}

func testGetLogs(t *testing.T, store *sqlstorage.Store) {
	require.NoError(t, store.Commit(context.Background(), tx1, tx2, tx3))

	cursor, err := store.GetLogs(context.Background(), ledger.NewLogsQuery())
	require.NoError(t, err)
	require.Equal(t, ledger.QueryDefaultPageSize, cursor.PageSize)

	require.Equal(t, 3, len(cursor.Data))
	require.Equal(t, uint64(2), cursor.Data[0].ID)
	require.Equal(t, tx3.Postings, cursor.Data[0].Data.(core.Transaction).Postings)
	require.Equal(t, tx3.Reference, cursor.Data[0].Data.(core.Transaction).Reference)
	require.Equal(t, tx3.Timestamp, cursor.Data[0].Data.(core.Transaction).Timestamp)

	cursor, err = store.GetLogs(context.Background(), &ledger.LogsQuery{
		PageSize: 1,
	})
	require.NoError(t, err)
	// Should get only the first log.
	require.Equal(t, 1, cursor.PageSize)
	require.Equal(t, uint64(2), cursor.Data[0].ID)

	cursor, err = store.GetLogs(context.Background(), &ledger.LogsQuery{
		AfterID:  cursor.Data[0].ID,
		PageSize: 1,
	})
	require.NoError(t, err)
	// Should get only the second log.
	require.Equal(t, 1, cursor.PageSize)
	require.Equal(t, uint64(1), cursor.Data[0].ID)

	cursor, err = store.GetLogs(context.Background(), &ledger.LogsQuery{
		Filters: ledger.LogsQueryFilters{
			StartTime: now.Add(-2 * time.Hour),
			EndTime:   now.Add(-1 * time.Hour),
		},
		PageSize: 10,
	})
	require.NoError(t, err)
	require.Equal(t, 10, cursor.PageSize)
	// Should get only the second log, as StartTime is inclusive and EndTime exclusive.
	require.Len(t, cursor.Data, 1)
	require.Equal(t, uint64(1), cursor.Data[0].ID)
}
