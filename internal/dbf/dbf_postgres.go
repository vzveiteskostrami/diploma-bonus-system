package dbf

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"time"

	_ "github.com/lib/pq"
	"github.com/vzveiteskostrami/diploma-bonus-system/internal/logging"
	"github.com/vzveiteskostrami/diploma-bonus-system/internal/misc"
)

type PGStorage struct {
	db *sql.DB
}

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

func (d *PGStorage) Register(login *string, password *string) (code int, err error) {
	code = http.StatusOK
	hashLogin := misc.Hash256(*login)
	rows, err := d.db.QueryContext(context.Background(), "SELECT 1 FROM UDATA WHERE USER_NAME=$1;", hashLogin)
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
		err = errors.New("логин уже занят")
		logging.S().Infoln(*login, ":", err)
		code = http.StatusConflict
		return
	}

	rows, err = d.db.QueryContext(context.Background(), "SELECT NEXTVAL('GEN_OID');")
	if err != nil || rows.Err() != nil {
		logging.S().Error(err)
		code = http.StatusInternalServerError
		return
	}

	var userId int64
	if rows.Next() {
		err = rows.Scan(&userId)
		if err != nil {
			logging.S().Error(err)
			code = http.StatusInternalServerError
			return
		}
	}

	hashPwd := misc.Hash256(*password)

	_, err = d.db.ExecContext(context.Background(), "INSERT INTO UDATA (USERID,USER_NAME,USER_PWD,DELETE_FLAG) VALUES ($1,$2,$3,false);", userId, hashLogin, hashPwd)
	if err != nil {
		logging.S().Error(err)
		code = http.StatusInternalServerError
		return
	}

	return
}

func (d *PGStorage) GetUserOrders(userID int64) (orders []Order, err error) {
	rows, err := d.db.QueryContext(context.Background(), "SELECT OID,USERID,NUMBER,STATUS,ACCRUAL,UPLOADED_AT,DELETE_FLAG from ORDERS WHERE USERID=$1 AND NOT DELETE_FLAG ORDER BY UPLOADED_AT;", userID)
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
		err = rows.Scan(&order.oid, &order.userid, &order.Number, &order.Status, &order.Accrual, &order.uploaded_at, &order.delete_flag)
		if err != nil {
			logging.S().Error()
			return
		}
		tm := order.uploaded_at.Format(time.RFC3339)
		order.Uploaded_at = &tm
		orders = append(orders, order)
	}
	return
}

func (d *PGStorage) GetUserBalance(userID int64) (balance Balance, err error) {
	rows, err := d.db.QueryContext(context.Background(), "SELECT SUM(ACCRUAL) as CURRENT,SUM(WITHDRAWN) as WITHDRAWN from ORDERS WHERE USERID=$1 AND NOT DELETE_FLAG;", userID)
	if err == nil && rows.Err() != nil {
		err = rows.Err()
	}
	if err != nil {
		logging.S().Error(err)
		return
	}
	defer rows.Close()

	balance = Balance{}
	for rows.Next() {
		err = rows.Scan(&balance.Current, &balance.Withdrawn)
		if err != nil {
			logging.S().Error()
			return
		}
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

	rows.Close()
	rows, err = d.db.QueryContext(context.Background(), "SELECT NEXTVAL('GEN_OID');")
	if err != nil || rows.Err() != nil {
		logging.S().Error(err)
		code = http.StatusInternalServerError
		return
	}

	var oId int64
	if rows.Next() {
		err = rows.Scan(&oId)
		if err != nil {
			logging.S().Error(err)
			code = http.StatusInternalServerError
			return
		}
	}

	_, err = d.db.ExecContext(context.Background(), "INSERT INTO ORDERS (OID,USERID,NUMBER,STATUS,ACCRUAL,UPLOADED_AT,DELETE_FLAG) VALUES ($1,$2,$3,'NEW',0,$4,false);", oId, userID, number, time.Now())
	if err != nil {
		logging.S().Error(err)
		code = http.StatusInternalServerError
	} else {
		code = http.StatusAccepted
	}

	return
}

func (d *PGStorage) Authent(login *string, password *string) (token string, code int, err error) {
	token = ""
	code = http.StatusOK
	hashLogin := misc.Hash256(*login)
	hashPwd := misc.Hash256(*password)
	rows, err := d.db.QueryContext(context.Background(), "SELECT USERID,USER_PWD FROM UDATA WHERE USER_NAME=$1;", hashLogin)
	if err == nil && rows.Err() != nil {
		err = rows.Err()
	}
	if err != nil {
		logging.S().Error(err)
		code = http.StatusInternalServerError
		return
	}
	defer rows.Close()

	ok := false
	var userId int64
	var dtbPwd string
	if rows.Next() {
		err = rows.Scan(&userId, &dtbPwd)
		if err != nil {
			logging.S().Error(err)
			code = http.StatusInternalServerError
			return
		}
		ok = true
	}

	if ok {
		ok = hashPwd == dtbPwd
	}

	if ok {
		token, err = misc.MakeToken(userId)
		if err != nil {
			logging.S().Error(err)
			code = http.StatusInternalServerError
		}
	} else {
		err = errors.New("неверная пара логин/пароль")
		logging.S().Infoln(*login, *password, ":", err)
		code = http.StatusUnauthorized
	}
	return
}
