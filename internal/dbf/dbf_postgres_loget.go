package dbf

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	_ "github.com/lib/pq"
	"github.com/vzveiteskostrami/diploma-bonus-system/internal/config"
	"github.com/vzveiteskostrami/diploma-bonus-system/internal/logging"
	"github.com/vzveiteskostrami/diploma-bonus-system/internal/misc"
)

func (d *PGStorage) OrdersCheck() {
	sql := "SELECT OID,NUMBER,ACCRUAL,STATUS from ORDERS WHERE NOT DELETE_FLAG AND STATUS IN (0,1);"
	rows, err := d.db.QueryContext(context.Background(), sql)
	if err == nil && rows.Err() != nil {
		err = rows.Err()
	}
	if err != nil {
		logging.S().Error(err)
		return
	}
	defer rows.Close()

	exec := ""
	params := []interface{}{}

	order := Order{}
	num := 1
	for rows.Next() {
		err = rows.Scan(&order.oid, &order.Number, &order.Accrual, &order.status)
		if err != nil {
			logging.S().Error()
			return
		}
		if order.Number != nil && *order.Number != "" {
			loy, ok := getOrderInfo(*order.Number)
			if ok {
				if *order.Accrual != *loy.Accrual || *order.status != *loy.status {
					params = append(params, order.oid, loy.Accrual, loy.status)
					delSQLBody += "($" + strconv.Itoa(num) + ",$" + strconv.Itoa(num+1) + ",$" + strconv.Itoa(num+2) + ")"
					num += 3
				}
			}
		}
	}

	if exec != "" {
		exec = "update orders set status=tmp.status,accrual=tmp.accrual from (values " +
			exec +
			") as tmp (oID,status,accrual) where orders.oID=tmp.oID;"

		_, err = d.db.ExecContext(context.Background(), exec, params...)
		if err != nil {
			logging.S().Infoln("SQL:", exec)
			logging.S().Infoln("Params:", params)
			logging.S().Infoln("Ошибка:", err)
		}
	}
}

type getOrder struct {
	Order   string
	Status  string
	Accural float32
}

func getOrderInfo(number string) (o Order, ok bool) {
	ok = false
	client := &http.Client{}
	r, err := client.Get(config.Addresses.Out.Host + ":" + strconv.Itoa(config.Addresses.Out.Port) + "/" + number)
	if err != nil {
		logging.S().Infoln("Get request error")
		logging.S().Infoln("Number:", number)
		logging.S().Infoln("Ошибка:", err)
		return
	}
	defer r.Body.Close()
	if r.StatusCode == http.StatusOK {
		getO := getOrder{}
		if err = json.NewDecoder(r.Body).Decode(&getO); err != nil {
			logging.S().Infoln("Get JSON error")
			logging.S().Infoln("Number:", number)
			logging.S().Infoln("Ошибка:", err)
			return
		}
		v := misc.StatusStrToInt(getO.Status)
		o.status = &v
		o.Accrual = &getO.Accural
		ok = true
	} else {
		logging.S().Infoln("Get answer no 200")
		logging.S().Infoln("Number:", number)
		logging.S().Infoln("Ошибка:", r.StatusCode, r.Status)
	}
	return
}
