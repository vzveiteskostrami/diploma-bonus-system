package dbf

import (
	"context"
	"fmt"
	"net/http"
	"time"

	_ "github.com/lib/pq"
	"github.com/vzveiteskostrami/diploma-bonus-system/internal/logging"
)

func (d *PGStorage) WithdrawAccrual(userID int64, number string, withdraw float32) (code int, err error) {
	code = http.StatusOK

	logging.S().Infoln("Записываю списание:", withdraw)

	var oID int64
	oID, err = d.nextOID()
	if err != nil {
		logging.S().Error(err)
		code = http.StatusInternalServerError
		return
	}

	//Max difference between 469.48697 and -220420.84 allowed is 0.01, but difference was
	_, err = d.db.ExecContext(context.Background(), "INSERT INTO ORDERS "+
		"(OID,USERID,NUMBER,STATUS,ACCRUAL,ACCRUAL_DATE,DELETE_FLAG) "+
		"VALUES "+
		"($1,$2,$3,0,$4,$5,false);",
		oID, userID, number, -withdraw, time.Now())
	if err != nil {
		logging.S().Error(err)
		code = http.StatusInternalServerError
	}

	uid := int64(0)
	num := ""
	accr := float32(0)
	accrd := time.Now()
	rows, err := d.db.QueryContext(context.Background(), "SELECT USERID,NUMBER,ACCRUAL,ACCRUAL_DATE from ORDERS")
	if err == nil && rows.Err() != nil {
		err = rows.Err()
	}
	if err != nil {
		logging.S().Error(err)
		return
	}
	defer rows.Close()

	fmt.Println("-----------------------------------------------------------------------")
	for rows.Next() {
		err = rows.Scan(&uid, &num, &accr, &accrd)
		fmt.Println(uid, num, accr, accrd)
	}
	fmt.Println("-----------------------------------------------------------------------")

	return
}

func (d *PGStorage) GetUserWithdraw(userID int64) (list []Withdraw, err error) {
	rows, err := d.db.QueryContext(context.Background(), "SELECT NUMBER,ACCRUAL,ACCRUAL_DATE from ORDERS WHERE USERID=$1 AND NOT DELETE_FLAG AND ACCRUAL < 0;", userID)
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
		*item.Sum = -*item.Sum
		list = append(list, item)
	}
	return
}
