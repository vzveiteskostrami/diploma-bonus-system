package dbf

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	_ "github.com/lib/pq"
	"github.com/vzveiteskostrami/diploma-bonus-system/internal/config"
	"github.com/vzveiteskostrami/diploma-bonus-system/internal/logging"
)

func (d *PGStorage) DBFInit() int64 {
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
	nextNumDB, err := d.tableInitData()
	if err != nil {
		logging.S().Panic(err)
	}
	return nextNumDB
}

func (d *PGStorage) DBFClose() {
	if d.db != nil {
		d.db.Close()
	}
}

func (d *PGStorage) tableInitData() (int64, error) {
	if d.db == nil {
		return -1, errors.New("база данных не инициализирована")
	}
	_, err := d.db.ExecContext(context.Background(),
		"CREATE TABLE IF NOT EXISTS UDATA(USERID bigint not null,USER_NAME character varying(64) NOT NULL,USER_PWD character varying(64) NOT NULL,DELETE_FLAG boolean DEFAULT false);"+
			"CREATE UNIQUE INDEX IF NOT EXISTS udata1 ON udata (USERID);"+
			"CREATE UNIQUE INDEX IF NOT EXISTS udata2 ON udata (USER_NAME);"+
			"CREATE TABLE IF NOT EXISTS ORDERS(OID bigint not null,USERID bigint not null,NUMBER character varying(64) NOT NULL,STATUS character varying(10) NOT NULL,ACCRUAL double precision NOT NULL,WITHDRAWN double precision NOT NULL, UPLOADED_AT timestamp with time zone NOT NULL,DELETE_FLAG boolean DEFAULT false);"+
			"CREATE UNIQUE INDEX IF NOT EXISTS orders1 ON orders (OID);"+
			"CREATE INDEX IF NOT EXISTS orders2 ON orders (UPLOADED_AT);"+
			"CREATE INDEX IF NOT EXISTS orders3 ON orders (USERID,UPLOADED_AT);"+
			"CREATE UNIQUE INDEX IF NOT EXISTS orders4 ON orders (NUMBER);"+
			"create sequence if not exists gen_oid as bigint minvalue 1 no maxvalue start 1 no cycle;")

	if err != nil {
		return -1, err
	}
	return 0, nil
}

func (d *PGStorage) PingDBf(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	if d.db == nil {
		http.Error(w, `База данных не открыта`, http.StatusInternalServerError)
		return
	}

	err := d.db.Ping()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (d *PGStorage) PrintDBF() {
	rows, err := d.db.QueryContext(context.Background(), "SELECT OWNERID,SHORTURL,ORIGINALURL from urlstore;")
	if err != nil {
		logging.S().Panic(err)
	}
	if rows.Err() != nil {
		logging.S().Panic(rows.Err())
	}
	defer rows.Close()

	var ow int64
	var sho string
	var fu string
	logging.S().Infow("--------------")
	for rows.Next() {
		err = rows.Scan(&ow, &sho, &fu)
		if err != nil {
			logging.S().Panic(err)
		}
		logging.S().Infow("", "owher", strconv.FormatInt(ow, 10), "short", sho, "full", fu)
	}
	logging.S().Infow("`````````````")
}
