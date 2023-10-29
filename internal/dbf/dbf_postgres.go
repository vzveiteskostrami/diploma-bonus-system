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

// После просмотра вебинара я в общем и целом понял, что имеется в виду под миграциями,
// но не понял как конкретно пользоваться пакетами. Чтение литературы в интернете ни к чему
// конкретному не привело. Где-то пишут, нужны утилиты, где-то, что вроде работает и без них.
// Но какой-то общей тенденции я не нашёл. Поэтому пусть создание базы останется так как есть.
// Имена индексов поменял.
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
		"CREATE UNIQUE INDEX IF NOT EXISTS udata_unique_on_userid ON udata (USERID);" +
		"CREATE INDEX IF NOT EXISTS udata_on_user_name ON udata (USER_NAME);"

	exec += "" +
		"CREATE TABLE IF NOT EXISTS ORDERS(" +
		"OID bigint not null," +
		"USERID bigint not null," +
		"NUMBER character varying(64) NOT NULL," +
		"STATUS smallint NOT NULL," +
		"ACCRUAL double precision NOT NULL," +
		"ACCRUAL_DATE timestamp with time zone NOT NULL," +
		"DELETE_FLAG boolean DEFAULT false" +
		");"
	exec += "" +
		"CREATE UNIQUE INDEX IF NOT EXISTS orders_unique_on_oid ON orders (OID);" +
		"CREATE INDEX IF NOT EXISTS orders_on_userid ON orders (USERID);" +
		"CREATE UNIQUE INDEX IF NOT EXISTS orders_unique_on_number ON orders (NUMBER);"

	/*
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
			"CREATE UNIQUE INDEX IF NOT EXISTS draws_unique_on_oid ON draws (OID);" +
			"CREATE INDEX IF NOT EXISTS draws_on_userid ON draws (USERID);"
	*/

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
