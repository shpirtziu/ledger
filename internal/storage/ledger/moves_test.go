//go:build it

package ledger

import (
	"database/sql"
	"fmt"
	"github.com/alitto/pond"
	"github.com/formancehq/go-libs/bun/bunpaginate"
	"github.com/formancehq/go-libs/logging"
	"github.com/formancehq/go-libs/platform/postgres"
	"github.com/formancehq/go-libs/time"
	ledger "github.com/formancehq/ledger/internal"
	ledgercontroller "github.com/formancehq/ledger/internal/controller/ledger"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"math/big"
	"math/rand"
	"testing"
)

func TestMovesInsert(t *testing.T) {
	t.Parallel()

	t.Run("nominal", func(t *testing.T) {
		t.Parallel()

		store := newLedgerStore(t)
		ctx := logging.TestingContext()

		tx := Transaction{}
		require.NoError(t, store.insertTransaction(ctx, &tx))

		account := &Account{}
		_, err := store.upsertAccount(ctx, account)
		require.NoError(t, err)

		now := time.Now()

		// we will insert 5 tx at five different timestamps
		// t0 ---------> t1 ---------> t2 ---------> t3 ----------> t4
		// m1 ---------> m3 ---------> m4 ---------> m2 ----------> m5
		t0 := now
		t1 := t0.Add(time.Hour)
		t2 := t1.Add(time.Hour)
		t3 := t2.Add(time.Hour)
		t4 := t3.Add(time.Hour)

		// insert a first tx at t0
		m1 := Move{
			Ledger:              store.ledger.Name,
			IsSource:            true,
			Account:             "world",
			AccountAddressArray: []string{"world"},
			Amount:              (*bunpaginate.BigInt)(big.NewInt(100)),
			Asset:               "USD",
			TransactionSeq:      tx.Seq,
			AccountSeq:          account.Seq,
			InsertionDate:       t0,
			EffectiveDate:       t0,
		}
		require.NoError(t, store.insertMoves(ctx, &m1))
		require.NotNil(t, m1.PostCommitVolumes)
		require.Equal(t, ledger.Volumes{
			Input:  big.NewInt(0),
			Output: big.NewInt(100),
		}, *m1.PostCommitVolumes)
		require.NotNil(t, m1.PostCommitEffectiveVolumes)
		require.Equal(t, ledger.Volumes{
			Input:  big.NewInt(0),
			Output: big.NewInt(100),
		}, *m1.PostCommitEffectiveVolumes)

		// add a second move at t3
		m2 := Move{
			Ledger:              store.ledger.Name,
			IsSource:            false,
			Account:             "world",
			AccountAddressArray: []string{"world"},
			Amount:              (*bunpaginate.BigInt)(big.NewInt(50)),
			Asset:               "USD",
			TransactionSeq:      tx.Seq,
			AccountSeq:          account.Seq,
			InsertionDate:       t3,
			EffectiveDate:       t3,
		}
		require.NoError(t, store.insertMoves(ctx, &m2))
		require.NotNil(t, m2.PostCommitVolumes)
		require.Equal(t, ledger.Volumes{
			Input:  big.NewInt(50),
			Output: big.NewInt(100),
		}, *m2.PostCommitVolumes)
		require.NotNil(t, m2.PostCommitEffectiveVolumes)
		require.Equal(t, ledger.Volumes{
			Input:  big.NewInt(50),
			Output: big.NewInt(100),
		}, *m2.PostCommitEffectiveVolumes)

		// add a third move at t1
		m3 := Move{
			Ledger:              store.ledger.Name,
			IsSource:            true,
			Account:             "world",
			AccountAddressArray: []string{"world"},
			Amount:              (*bunpaginate.BigInt)(big.NewInt(200)),
			Asset:               "USD",
			TransactionSeq:      tx.Seq,
			AccountSeq:          account.Seq,
			InsertionDate:       t1,
			EffectiveDate:       t1,
		}
		require.NoError(t, store.insertMoves(ctx, &m3))
		require.NotNil(t, m3.PostCommitVolumes)
		require.Equal(t, ledger.Volumes{
			Input:  big.NewInt(50),
			Output: big.NewInt(300),
		}, *m3.PostCommitVolumes)
		require.NotNil(t, m3.PostCommitEffectiveVolumes)
		require.Equal(t, ledger.Volumes{
			Input:  big.NewInt(0),
			Output: big.NewInt(300),
		}, *m3.PostCommitEffectiveVolumes)

		// add a fourth move at t2
		m4 := Move{
			Ledger:              store.ledger.Name,
			IsSource:            false,
			Account:             "world",
			AccountAddressArray: []string{"world"},
			Amount:              (*bunpaginate.BigInt)(big.NewInt(50)),
			Asset:               "USD",
			TransactionSeq:      tx.Seq,
			AccountSeq:          account.Seq,
			InsertionDate:       t2,
			EffectiveDate:       t2,
		}
		require.NoError(t, store.insertMoves(ctx, &m4))
		require.NotNil(t, m4.PostCommitVolumes)
		require.Equal(t, ledger.Volumes{
			Input:  big.NewInt(100),
			Output: big.NewInt(300),
		}, *m4.PostCommitVolumes)
		require.NotNil(t, m4.PostCommitEffectiveVolumes)
		require.Equal(t, ledger.Volumes{
			Input:  big.NewInt(50),
			Output: big.NewInt(300),
		}, *m4.PostCommitEffectiveVolumes)

		// add a fifth move at t4
		m5 := Move{
			Ledger:              store.ledger.Name,
			IsSource:            false,
			Account:             "world",
			AccountAddressArray: []string{"world"},
			Amount:              (*bunpaginate.BigInt)(big.NewInt(50)),
			Asset:               "USD",
			TransactionSeq:      tx.Seq,
			AccountSeq:          account.Seq,
			InsertionDate:       t4,
			EffectiveDate:       t4,
		}
		require.NoError(t, store.insertMoves(ctx, &m5))
		require.NotNil(t, m5.PostCommitVolumes)
		require.Equal(t, ledger.Volumes{
			Input:  big.NewInt(150),
			Output: big.NewInt(300),
		}, *m5.PostCommitVolumes)
		require.NotNil(t, m5.PostCommitEffectiveVolumes)
		require.Equal(t, ledger.Volumes{
			Input:  big.NewInt(150),
			Output: big.NewInt(300),
		}, *m5.PostCommitEffectiveVolumes)
	})

	t.Run("with high concurrency", func(t *testing.T) {
		t.Parallel()

		store := newLedgerStore(t)
		ctx := logging.TestingContext()

		wp := pond.New(10, 10)
		for i := 0; i < 1000; i++ {
			wp.Submit(func() {
				for {
					sqlTx, err := store.GetDB().BeginTx(ctx, &sql.TxOptions{})
					require.NoError(t, err)
					storeCP := store.WithDB(sqlTx)

					src := fmt.Sprintf("accounts:%d", rand.Intn(1000000))
					dst := fmt.Sprintf("accounts:%d", rand.Intn(1000000))

					tx := ledger.NewTransaction().WithPostings(
						ledger.NewPosting(src, dst, "USD", big.NewInt(1)),
					)
					err = storeCP.CommitTransaction(ctx, &tx)
					if errors.Is(err, postgres.ErrDeadlockDetected) {
						require.NoError(t, sqlTx.Rollback())
						continue
					}
					require.NoError(t, err)
					require.NoError(t, sqlTx.Commit())
					return
				}
			})
		}
		wp.StopAndWait()

		aggregatedBalances, err := store.GetAggregatedBalances(ctx, ledgercontroller.NewGetAggregatedBalancesQuery(ledgercontroller.PITFilter{}, nil, true))
		require.NoError(t, err)
		RequireEqual(t, ledger.BalancesByAssets{
			"USD": big.NewInt(0),
		}, aggregatedBalances)
	})
}
