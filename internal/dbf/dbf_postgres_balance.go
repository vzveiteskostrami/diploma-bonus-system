package dbf

import (
	"context"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/vzveiteskostrami/diploma-bonus-system/internal/logging"
)

func (d *PGStorage) GetUserBalance(userID int64) (balance Balance, err error) {
	exec := "" +
		"SELECT " +
		"SUM(CASE WHEN ACCRUAL > 0 THEN ACCRUAL ELSE 0 END)," +
		"SUM(CASE WHEN ACCRUAL < 0 THEN ACCRUAL ELSE 0 END) " +
		"FROM ORDERS " +
		"WHERE USERID=$1 AND NOT DELETE_FLAG;"

	rows, err := d.db.QueryContext(context.Background(), exec, userID)
	if err == nil && rows.Err() != nil {
		err = rows.Err()
	}
	if err != nil {
		logging.S().Infoln("SQL: " + exec)
		logging.S().Error(err)
		return
	}
	defer rows.Close()

	balance = Balance{}
	if rows.Next() {
		err = rows.Scan(&balance.Current, &balance.Withdrawn)
		if err != nil {
			logging.S().Error(err)
			return
		}
	}
	rows.Close()

	if balance.Current == nil {
		balance.Current = new(float32)
	}
	if balance.Withdrawn == nil {
		balance.Withdrawn = new(float32)
	}

	fmt.Println("Получили баланс", *balance.Current, *balance.Withdrawn)

	*balance.Current += *balance.Withdrawn
	*balance.Withdrawn = -*balance.Withdrawn

	fmt.Println("Вернули баланс", *balance.Current, *balance.Withdrawn)

	d.DrawTable()
	return
}
