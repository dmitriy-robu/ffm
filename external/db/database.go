package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
)

type SqlInterface interface {
	DoInTransaction(do func(tx *sql.Tx) error) error
	Begin() error
	Commit() error
	Rollback() error
	GetTx() *sql.Tx
	GetExecer() QueryExecer
}

type QueryExecer interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
}

type DataBase struct {
	db *sql.DB
	tx *sql.Tx
}

func NewMysql(
	db *sql.DB,
) *DataBase {
	return &DataBase{
		db: db,
	}
}

func (u *DataBase) Begin() error {
	var err error
	u.tx, err = u.db.Begin()
	return err
}

func (u *DataBase) DoInTransaction(do func(tx *sql.Tx) error) (err error) {
	tx, err := u.db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				err = rollbackErr
				return
			}
			panic(p)
		} else if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				err = rollbackErr
				return
			}
		} else {
			err = tx.Commit()
		}
	}()

	err = do(tx)
	return err
}

func (u *DataBase) Commit() error {
	if u.tx == nil {
		return nil
	}

	err := u.tx.Commit()
	if err != nil {
		return err
	}

	u.tx = nil
	return err
}

func (u *DataBase) Rollback() error {
	if u.tx == nil {
		return fmt.Errorf("transaction is not started: %w", errors.New("rollback"))
	}

	err := u.tx.Rollback()
	if err != nil {
		return err
	}

	u.tx = nil
	return nil
}

func (u *DataBase) GetDB() *sql.DB {
	return u.db
}

func (u *DataBase) GetTx() *sql.Tx {
	return u.tx
}

func (u *DataBase) GetExecer() QueryExecer {
	if u.tx != nil {
		return u.tx
	}
	return u.db
}
