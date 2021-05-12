package app

import (
	"context"
	"errors"
	"sync"

	"github.com/ds248a/air/config"
	"github.com/jackc/pgx/v4/pgxpool"
)

const (
	pg_db       = "db_point"
	pg_user     = "point"
	pg_password = "SeCreT10"
)

var (
	errPostgreConfig = errors.New("Postgre not configured")
)

// --------------------------------
//    Postgre
// --------------------------------

func newPostgre(cfg *config.Postgre) error {
	if cfg == nil {
		return errPostgreConfig
	}

	if !cfg.Enable {
		return nil
	}

	err := cfg.Options(pg_db, pg_user, pg_password)
	if err != nil {
		return err
	}

	pdb, err = pgxpool.Connect(context.Background(), cfg.Dsn)
	if err != nil {
		return err
	}

	callOnExit(func(ctx context.Context, wg *sync.WaitGroup) {
		defer wg.Done()

		dst := make(chan bool)
		go func(db *pgxpool.Pool) {
			db.Close()
			Debug("-- Postgre Close")
			dst <- true
		}(pdb)

	loop:
		for {
			select {
			case <-ctx.Done():
				break loop
			case <-dst:
				break loop
			}
		}
	})

	return nil
}

func Postgre() *pgxpool.Pool {
	return pdb
}
