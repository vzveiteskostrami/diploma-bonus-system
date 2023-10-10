package dbf

import (
	"context"
	"database/sql"
	"errors"

	_ "github.com/lib/pq"
	"github.com/vzveiteskostrami/diploma-bonus-system/internal/config"
	"github.com/vzveiteskostrami/diploma-bonus-system/internal/logging"
)

type PGStorage struct {
	db *sql.DB
}

func (d *PGStorage) tableInitData() error {
	if d.db == nil {
		return errors.New("база данных не инициализирована")
	}

	exec := "" +
		"CREATE TABLE IF NOT EXISTS UDATA(" +
		"USERID bigint not null," +
		"USER_NAME character varying(64) NOT NULL," +
		"USER_PWD character varying(64) NOT NULL," +
		"DELETE_FLAG boolean DEFAULT false" +
		");"
	exec += "" +
		"CREATE UNIQUE INDEX IF NOT EXISTS udata1 ON udata (USERID);" +
		"CREATE UNIQUE INDEX IF NOT EXISTS udata2 ON udata (USER_NAME);"

	exec += "" +
		"CREATE TABLE IF NOT EXISTS ORDERS(" +
		"OID bigint not null," +
		"USERID bigint not null," +
		"NUMBER character varying(64) NOT NULL," +
		"STATUS smallint NOT NULL," +
		"ACCRUAL double precision NOT NULL," +
		"NEW_DATE timestamp with time zone NOT NULL," +
		"DELETE_FLAG boolean DEFAULT false" +
		");"
	exec += "" +
		"CREATE UNIQUE INDEX IF NOT EXISTS orders1 ON orders (OID);" +
		"CREATE INDEX IF NOT EXISTS orders2 ON orders (USERID);" +
		"CREATE UNIQUE INDEX IF NOT EXISTS orders3 ON orders (NUMBER);"

	exec += "" +
		"CREATE TABLE IF NOT EXISTS DRAWS(" +
		"OID bigint not null," +
		"USERID bigint not null," +
		"NUMBER character varying(64) NOT NULL," +
		"WITHDRAW double precision NOT NULL," +
		"WITHDRAW_DATE timestamp with time zone NOT NULL," +
		"DELETE_FLAG boolean DEFAULT false" +
		");"
	exec += "" +
		"CREATE UNIQUE INDEX IF NOT EXISTS draws1 ON draws (OID);" +
		"CREATE INDEX IF NOT EXISTS draws2 ON draws (USERID);"

	exec += "" +
		"create sequence if not exists gen_oid as bigint minvalue 1 no maxvalue start 1 no cycle;"

	_, err := d.db.ExecContext(context.Background(), exec)
	if err != nil {
		return err
	}
	return nil
}

func (d *PGStorage) Init() {
	var err error

	d.db, err = sql.Open("postgres", config.Storage.DBConnect)
	if err != nil {
		logging.S().Panic(err)
	}
	logging.S().Infof("Объявлено соединение с %s", config.Storage.DBConnect)

	err = d.db.Ping()
	if err != nil {
		logging.S().Panic(err)
	}
	logging.S().Infof("Установлено соединение с %s", config.Storage.DBConnect)
	err = d.tableInitData()
	if err != nil {
		logging.S().Panic(err)
	}
}

func (d *PGStorage) Close() {
	if d.db != nil {
		d.db.Close()
	}
}
