package dbf

import (
	"context"
	"strconv"

	_ "github.com/lib/pq"
	"github.com/vzveiteskostrami/diploma-bonus-system/internal/logging"
)

// Изучив схему хранения данных в одной таблице я пришёл к выводу, что это неудобно по нескольким причинам.
// И сомнительно удобно только по одной, а именно - запрос баланса делается по одной таблице. Сомнительно, потому
// что я не понимаю, почему обращение к двум таблицам считается катастрофой, которую надо исправлять.
// Неудобно, потому что:
//
//  1. У списаниий нет статуса. У начислений есть. То есть, структуры разные. Что писать в STATUS при списании?
//
//  2. В списаниях номер заказа не уникален, так как с одного заказа баллы могут списываться много раз.
//     В начислениях номер заказа уникален. Об этом говорят две фразы:
//     a) 200 — номер заказа уже был загружен этим пользователем
//     б) 409 — номер заказа уже был загружен другим пользователем
//     Если убрать уникальный индекс и делать проверки, отличая в одной таблице начисления от списаний,
//     то как минимум в этом месте громоздкость будет выглядеть гораздо хуже, чем обращение к двум таблицам
//     при получении баланса.
//
//  3. Так как довольно часто в аргументах фигурирует "а что будет если", я тоже применю сей приём.
//     А что будет, если мы захотим сделать сущности возврат списаний и возврат начислений? А в реальной жизни
//     без оговорки данных сущностей проект даже не начнётся. Тогда возврат начисления
//     будет выглядеть как начисление с минусом, а возврат списания как списание с минусом. При конструкции "одна таблица"
//     выражение этих сущностей будет уже занято и придётся придумывать что-то вместо.
//
//  4. Одна таблица будет подвергнута бОльшей нагрузке и будет разрастаться быстрее. Соответственно, будет более
//     тяжела в обслуживании.
//
//  5. По одной таблице все запросы начинают выглядеть более громоздко и менее читабельно из-за того, что к любым
//     операциям, кроме получения баланса, начинают добавляться проверки начисление это или удержание.
//
//     Возможно, если бы я не гипотетически рассмотрел переход с двух таблиц на одну, я бы нашёл ещё недостатки.
//     Но на этапе рассмотрения и размышлений это всё. Но, как мне кажется, этого достаточно, чтобы оставить всё как есть.
//     Я оставляю две таблицы, но переделываю их на один запрос.
func (d *PGStorage) GetUserBalance(userID int64) (balance Balance, err error) {
	exec := "" +
		"SELECT 1 as A,SUM(ACCRUAL) as CURRENT from ORDERS WHERE USERID=" + strconv.FormatInt(userID, 10) + " AND NOT DELETE_FLAG " +
		"UNION " +
		"SELECT 2 as A,SUM(WITHDRAW) as CURRENT from DRAWS WHERE USERID=" + strconv.FormatInt(userID, 10) + " AND NOT DELETE_FLAG " +
		"ORDER BY 1;"

		//		SELECT 1 as A,SUM(ACCRUAL) as CURRENT from ORDERS WHERE USERID=$1 AND NOT DELETE_FLAG
		//		UNION
	//SELECT 2 as A,SUM(WITHDRAW) as CURRENT from DRAWS WHERE USERID=$2 AND NOT DELETE_FLAG
	//ORDER BY 1;

	rows, err := d.db.QueryContext(context.Background(), exec, userID, userID)
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
	num := int64(0)
	if rows.Next() {
		err = rows.Scan(&num, &balance.Current)
		if err != nil {
			logging.S().Error(err)
			return
		}
	}
	if rows.Next() {
		err = rows.Scan(&num, &balance.Withdrawn)
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

	*balance.Current -= *balance.Withdrawn

	return
}
