package dbf

import (
	"context"
	"net/http"
	"time"

	_ "github.com/lib/pq"
	"github.com/vzveiteskostrami/diploma-bonus-system/internal/logging"
)

func (d *PGStorage) WithdrawAccrual(userID int64, number string, withdraw float32) (code int, err error) {
	code = http.StatusOK

	var oID int64
	oID, err = d.nextOID()
	if err != nil {
		logging.S().Error(err)
		code = http.StatusInternalServerError
		return
	}

	_, err = d.db.ExecContext(context.Background(), "INSERT INTO DRAWS (OID,USERID,NUMBER,WITHDRAW,WITHDRAW_DATE,DELETE_FLAG) VALUES ($1,$2,$3,$4,$5,false);", oID, userID, number, withdraw, time.Now())
	if err != nil {
		logging.S().Error(err)
		code = http.StatusInternalServerError
	}
	return
}

func (d *PGStorage) GetUserWithdraw(userID int64) (list []Withdraw, err error) {
	rows, err := d.db.QueryContext(context.Background(), "SELECT NUMBER,WITHDRAW,WITHDRAW_DATE from DRAWS WHERE USERID=$1 AND NOT DELETE_FLAG;", userID) // AND WITHDRAWN > 0
	if err == nil && rows.Err() != nil {
		err = rows.Err()
	}
	if err != nil {
		logging.S().Error(err)
		return
	}
	defer rows.Close()

	list = make([]Withdraw, 0)
	item := Withdraw{}
	for rows.Next() {
		err = rows.Scan(&item.Order, &item.Sum, &item.withdrawDate)
		if err != nil {
			logging.S().Error(err)
			return
		}
		if item.withdrawDate != nil {
			tm := item.withdrawDate.Format(time.RFC3339)
			item.ProcessedAt = &tm
		}
		list = append(list, item)
	}
	return
}
