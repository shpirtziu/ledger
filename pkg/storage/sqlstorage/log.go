package sqlstorage

import (
	"context"
	"database/sql"
	"fmt"
	json "github.com/gibson042/canonicaljson-go"
	"github.com/huandu/go-sqlbuilder"
	"github.com/numary/ledger/pkg/core"
	"github.com/pkg/errors"
	"time"
)

func (s *Store) appendLog(ctx context.Context, exec executor, log ...core.Log) error {
	var (
		sql  string
		args []interface{}
	)

	switch s.Schema().Flavor() {
	case sqlbuilder.SQLite:
		ib := sqlbuilder.NewInsertBuilder()
		ib.InsertInto(s.schema.Table("log"))
		ib.Cols("id", "type", "hash", "date", "data")
		for _, l := range log {
			data, err := json.Marshal(l.Data)
			if err != nil {
				panic(err)
			}

			ib.Values(l.ID, l.Type, l.Hash, l.Date.Format(time.RFC3339Nano), string(data))
		}
		sql, args = ib.BuildWithFlavor(s.schema.Flavor())
	case sqlbuilder.PostgreSQL:

		ids := make([]uint64, len(log))
		types := make([]string, len(log))
		hashs := make([]string, len(log))
		dates := make([]time.Time, len(log))
		datas := make([][]byte, len(log))

		for i, l := range log {
			data, err := json.Marshal(l.Data)
			if err != nil {
				panic(err)
			}

			ids[i] = l.ID
			types[i] = l.Type
			hashs[i] = l.Hash
			dates[i] = l.Date
			datas[i] = data
		}

		sql = fmt.Sprintf(`INSERT INTO "%s".log (id, type, hash, date, data) (SELECT * FROM unnest($1::int[], $2::varchar[], $3::varchar[], $4::timestamptz[], $5::jsonb[]))`, s.schema.Name())
		args = []interface{}{
			ids, types, hashs, dates, datas,
		}
	}

	_, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return s.error(err)
	}
	return nil
}

func (s *Store) AppendLog(ctx context.Context, logs ...core.Log) error {
	tx, err := s.schema.BeginTx(ctx, nil)
	if err != nil {
		return s.error(err)
	}
	defer tx.Rollback()

	err = s.appendLog(ctx, tx, logs...)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return s.error(err)
	}
	return nil
}

func (s *Store) lastLog(ctx context.Context, exec executor) (*core.Log, error) {

	sb := sqlbuilder.NewSelectBuilder()
	sb.From(s.schema.Table("log"))
	sb.Select("id", "type", "hash", "date", "data")
	sb.OrderBy("id desc")
	sb.Limit(1)

	sqlq, _ := sb.BuildWithFlavor(s.schema.Flavor())
	row := exec.QueryRowContext(ctx, sqlq)
	l := core.Log{}
	var (
		ts   sql.NullString
		data sql.NullString
	)
	err := row.Scan(&l.ID, &l.Type, &l.Hash, &ts, &data)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	t, err := time.Parse(time.RFC3339Nano, ts.String)
	if err != nil {
		return nil, err
	}
	l.Date = t.UTC()
	l.Data, err = core.HydrateLog(l.Type, data.String)
	if err != nil {
		return nil, err
	}
	return &l, nil
}

func (s *Store) LastLog(ctx context.Context) (*core.Log, error) {
	return s.lastLog(ctx, s.schema)
}

func (s *Store) logs(ctx context.Context, exec executor) ([]core.Log, error) {
	sb := sqlbuilder.NewSelectBuilder()
	sb.From(s.schema.Table("log"))
	sb.Select("id", "type", "hash", "date", "data")
	sb.OrderBy("id desc")

	sqlq, _ := sb.BuildWithFlavor(s.schema.Flavor())
	rows, err := exec.QueryContext(ctx, sqlq)
	if err != nil {
		return nil, s.error(err)
	}
	defer rows.Close()

	ret := make([]core.Log, 0)
	for rows.Next() {
		l := core.Log{}
		var (
			ts   sql.NullString
			data sql.NullString
		)

		err := rows.Scan(&l.ID, &l.Type, &l.Hash, &ts, &data)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, nil
			}
			return nil, err
		}
		t, err := time.Parse(time.RFC3339Nano, ts.String)
		if err != nil {
			return nil, err
		}
		l.Date = t.UTC()
		l.Data, err = core.HydrateLog(l.Type, data.String)
		if err != nil {
			return nil, errors.Wrap(err, "hydrating log")
		}
		ret = append(ret, l)
	}
	if rows.Err() != nil {
		return nil, s.error(rows.Err())
	}

	return ret, nil
}

func (s *Store) Logs(ctx context.Context) ([]core.Log, error) {
	return s.logs(ctx, s.schema)
}