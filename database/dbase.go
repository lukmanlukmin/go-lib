package database

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

type (
	ctxKey string

	SQLExec interface {
		sqlx.Execer
		sqlx.ExecerContext
		NamedExec(query string, arg interface{}) (sql.Result, error)
		NamedExecContext(ctx context.Context, query string, arg interface{}) (sql.Result, error)
	}
	SQLQuery interface {
		sqlx.Queryer
		GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
		SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
		PrepareNamedContext(ctx context.Context, query string) (*sqlx.NamedStmt, error)
	}

	SQLQueryExec interface {
		SQLExec
		SQLQuery
		Rebind(query string) string
	}

	WrapTransactionFunc func(ctx context.Context) error
)

// BeginTransaction wrapper function to handle transaction process v2
func BeginTransaction(ctx context.Context, db *sqlx.DB, fn WrapTransactionFunc, isolations ...sql.IsolationLevel) error {
	isolationLevel := sql.LevelRepeatableRead
	if len(isolations) > 0 {
		isolationLevel = isolations[0]
	}

	if tx := GetTxFromContext(ctx); tx != nil {
		return fn(ctx)
	}

	tx, err := db.BeginTxx(ctx, &sql.TxOptions{Isolation: isolationLevel})
	if err != nil {
		return err
	}

	ctx = context.WithValue(ctx, txKey, tx)

	defer func() {
		if p := recover(); p != nil {
			handleRollback(tx, "failed rollback when panic")
			panic(p)
		} else if err != nil {
			handleRollback(tx, "failed rollback on error")
		} else {
			handleCommit(tx)
		}
	}()

	err = fn(ctx)
	return err
}

const txKey ctxKey = "IsTransaction"

func GetTxFromContext(ctx context.Context) *sqlx.Tx {
	if tx, ok := ctx.Value(txKey).(*sqlx.Tx); ok {
		return tx
	}
	return nil
}

func handleRollback(tx *sqlx.Tx, msg string) {
	if err := tx.Rollback(); err != nil {
		logrus.WithError(err).Warn(msg)
	}
}

func handleCommit(tx *sqlx.Tx) {
	if err := tx.Commit(); err != nil {
		logrus.WithError(err).Warn("failed to commit transaction")
	}
}
