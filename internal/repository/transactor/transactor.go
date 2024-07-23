package transactor

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type QueryEngine interface {
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
}

type QueryEngineProvider interface {
	GetQueryEngine(ctx context.Context) QueryEngine // tx OR pool
}

type TransactionManager struct {
	Pool *pgxpool.Pool
}

func NewTransactionManager(pool *pgxpool.Pool) *TransactionManager {
	return &TransactionManager{Pool: pool}
}

type contextKey string

const transactionKey contextKey = "tx"

func (tm *TransactionManager) RunRepeatebleRead(ctx context.Context, fx func(ctxTX context.Context) error) error {
	tx, err := tm.Pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.RepeatableRead,
		AccessMode: pgx.ReadWrite,
	})
	if err != nil {
		return errors.New("Ошибка при попытке начать транзакцию: " + err.Error())
	}

	if err := fx(context.WithValue(ctx, "tx", tx)); err != nil {
		return errors.New("Ошибка при выполнении транзакции: " + tx.Rollback(ctx).Error())
	}

	if err := tx.Commit(ctx); err != nil {
		return errors.New("Ошибка при выполнении транзакции: " + tx.Rollback(ctx).Error())
	}

	return nil
}

func (tm *TransactionManager) GetQueryEngine(ctx context.Context) QueryEngine {
	tx, ok := ctx.Value(transactionKey).(QueryEngine)
	if ok && tx != nil {
		return tx
	}

	return tm.Pool
}
