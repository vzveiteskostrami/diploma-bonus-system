package routes

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/vzveiteskostrami/diploma-bonus-system/internal/auth"
	"github.com/vzveiteskostrami/diploma-bonus-system/internal/dbf"
	"github.com/vzveiteskostrami/diploma-bonus-system/internal/logging"
	"github.com/vzveiteskostrami/diploma-bonus-system/internal/misc"
)

func OrdersPostf(w http.ResponseWriter, r *http.Request) {
	b, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	sOrder := string(b)
	if sOrder == "" || !misc.CheckLuhn(sOrder) {
		s := `Неверный формат номера заказа`
		http.Error(w, s, http.StatusUnprocessableEntity)
		logging.S().Infoln(s, ":", sOrder)
		return
	}

	logging.S().Infoln("ORDERIN:", sOrder)
	code, err := dbf.Store.SaveOrderNum(r.Context().Value(auth.CPuserID).(int64), sOrder)
	if err != nil {
		http.Error(w, err.Error(), code)
	} else {
		logging.S().Infoln("ORDERINOK:", sOrder)
		w.WriteHeader(code)
	}
}

func OrdersGetf(w http.ResponseWriter, r *http.Request) {
	completed := make(chan struct{})

	var orders []dbf.Order
	var err error

	go func() {
		orders, err = dbf.Store.GetUserOrders(r.Context().Value(auth.CPuserID).(int64))
		completed <- struct{}{}
	}()

	select {
	case <-completed:
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			if len(orders) == 0 {
				http.Error(w, "Нет данных для ответа", http.StatusNoContent)
				//w.Write([]byte("{}"))
			} else {
				var buf bytes.Buffer
				if err := json.NewEncoder(&buf).Encode(orders); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write(buf.Bytes())
				logging.S().Infoln("ORDERSGET:", buf.String())
			}
		}
	case <-r.Context().Done():
		logging.S().Infoln("Получение данных прервано на клиентской стороне")
		w.WriteHeader(http.StatusGone)
	}
}
