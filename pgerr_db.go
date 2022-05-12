package main

import (
	"context"

	"github.com/ds248a/nami/log"
	"github.com/jackc/pgx/v4/pgxpool"
)

var (
	sqlErrNoRows = "no rows in result set"
)

// --------------------------------
//    Postgres Error Model
// --------------------------------

// Формирование сообщения в случае ошибки выполнения SQL запроса.
func DbErr(err error) {
	if err != nil {
		if err.Error() == sqlErrNoRows {
			return
		}
		log.Err(err)
	}
}

// В случае ошибки SQL запроса выполняется отмена транзакции,
// с последующим формированием сообщения о ошибке.
func DbTxErr(ctx context.Context, tx *pgxpool.Tx, err error) {
	if err != nil {
		if err.Error() == sqlErrNoRows {
			return
		}
		DbTxRollback(ctx, tx)
		log.Err(err)
	}
}

// Завершает текущую транзакцию или возвращает сообщение о ошибке.
func DbTxCommit(ctx context.Context, tx *pgxpool.Tx) {
	if err := tx.Commit(ctx); err != nil {
		log.Err(err)
	}
}

// Отменяет текущую транзакцию SQL запроса.
func DbTxRollback(ctx context.Context, tx *pgxpool.Tx) {
	if err := tx.Rollback(ctx); err != nil {
		if err.Error() != `tx is closed` {
			log.Err(err)
		}
	}
}

// Определяет, привел ли SQL запрос к появлению ошибки.
func IsDbErr(err error) bool {
	if err != nil {
		if err.Error() == sqlErrNoRows {
			return false
		}
		return true
	}
	return false
}
