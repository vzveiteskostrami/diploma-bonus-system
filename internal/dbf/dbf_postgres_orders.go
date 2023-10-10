package dbf

import (
	"context"
	"errors"
	"net/http"
	"time"

	_ "github.com/lib/pq"
	"github.com/vzveiteskostrami/diploma-bonus-system/internal/logging"
	"github.com/vzveiteskostrami/diploma-bonus-system/internal/misc"
)

func (d *PGStorage) GetUserOrders(userID int64) (orders []Order, err error) {
	rows, err := d.db.QueryContext(context.Background(), "SELECT OID,USERID,NUMBER,STATUS,ACCRUAL,NEW_DATE,DELETE_FLAG "+
		"from ORDERS "+
		"WHERE USERID=$1 AND NOT DELETE_FLAG ORDER BY NEW_DATE;", userID)
	if err == nil && rows.Err() != nil {
		err = rows.Err()
	}
	if err != nil {
		logging.S().Error(err)
		return
	}
	defer rows.Close()

	orders = make([]Order, 0)
	order := Order{}
	for rows.Next() {
		err = rows.Scan(&order.oid, &order.userid, &order.Number, &order.status, &order.Accrual, &order.uploadedAt, &order.deleteFlag)
		if err != nil {
			logging.S().Error()
			return
		}
		status := misc.StatusIntToStr(order.status)
		tm := order.uploadedAt.Format(time.RFC3339)
		order.UploadedAt = &tm
		order.Status = &status
		orders = append(orders, order)
	}
	return
}

func (d *PGStorage) SaveOrderNum(userID int64, number string) (code int, err error) {
	lockWrite.Lock()
	defer lockWrite.Unlock()

	code = http.StatusOK
	rows, err := d.db.QueryContext(context.Background(), "SELECT USERID FROM ORDERS WHERE NUMBER=$1;", number)
	if err == nil && rows.Err() != nil {
		err = rows.Err()
	}
	if err != nil {
		logging.S().Error(err)
		code = http.StatusInternalServerError
		return
	}
	defer rows.Close()

	if rows.Next() {
		dbUserID := int64(0)
		rows.Scan(&dbUserID)
		if err != nil {
			logging.S().Error(err)
			code = http.StatusInternalServerError
			return
		}
		if dbUserID == userID {
			err = errors.New("номер заказа уже был загружен этим пользователем")
		} else {
			err = errors.New("номер заказа уже был загружен другим пользователем")
			code = http.StatusConflict
		}
		logging.S().Infoln(number, ":", err)
		return
	}

	var oID int64
	oID, err = d.nextOID()
	if err != nil {
		logging.S().Error(err)
		code = http.StatusInternalServerError
		return
	}

	_, err = d.db.ExecContext(context.Background(), "INSERT INTO ORDERS (OID,USERID,NUMBER,STATUS,ACCRUAL,WITHDRAWN,NEW_DATE,DELETE_FLAG) VALUES ($1,$2,$3,0,0,0,$4,false);", oID, userID, number, time.Now())
	if err != nil {
		logging.S().Error(err)
		code = http.StatusInternalServerError
	} else {
		code = http.StatusAccepted
	}

	return
}
