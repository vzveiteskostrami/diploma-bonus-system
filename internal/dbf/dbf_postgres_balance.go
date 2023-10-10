package dbf

import (
	"context"

	_ "github.com/lib/pq"
	"github.com/vzveiteskostrami/diploma-bonus-system/internal/logging"
)

func (d *PGStorage) GetUserBalance(userID int64) (balance Balance, err error) {
	rows, err := d.db.QueryContext(context.Background(), "SELECT SUM(ACCRUAL) as CURRENT from ORDERS WHERE USERID=$1 AND NOT DELETE_FLAG;", userID)
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
		err = rows.Scan(&balance.Current)
		if err != nil {
			logging.S().Error(err)
			return
		}
	}
	rows.Close()

	rows, err = d.db.QueryContext(context.Background(), "SELECT SUM(WITHDRAW) as WITHDRAW from DRAWS WHERE USERID=$1 AND NOT DELETE_FLAG;", userID)
	if err == nil && rows.Err() != nil {
		err = rows.Err()
	}
	if err != nil {
		logging.S().Error(err)
		return
	}
	for rows.Next() {
		err = rows.Scan(&balance.Withdrawn)
		if err != nil {
			logging.S().Error(err)
			return
		}
	}

	balance.Current -= balance.Withdrawn

	return
}
