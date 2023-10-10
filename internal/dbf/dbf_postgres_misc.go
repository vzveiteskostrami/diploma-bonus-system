package dbf

import (
	"context"
	"errors"
	"strconv"

	_ "github.com/lib/pq"
	"github.com/vzveiteskostrami/diploma-bonus-system/internal/logging"
)

func (d *PGStorage) UserIDExists(userID int64) (ok bool, err error) {
	ok = false
	rows, err := d.db.QueryContext(context.Background(), "SELECT 1 FROM UDATA WHERE USERID=$1;", userID)
	if err == nil && rows.Err() != nil {
		err = rows.Err()
	}
	if err != nil {
		logging.S().Error(err)
		return
	}
	defer rows.Close()
	ok = rows.Next()
	return
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

func (d *PGStorage) nextOID() (oid int64, err error) {
	rows, err := d.db.QueryContext(context.Background(), "SELECT NEXTVAL('GEN_OID');")
	if err == nil || rows.Err() != nil {
		err = rows.Err()
	}
	if err != nil {
		logging.S().Error(err)
		return
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(&oid)
		if err != nil {
			logging.S().Error(err)
		}
	} else {
		err = errors.New("не вышло получить значение счётчика")
		logging.S().Error(err)
	}
	return
}
