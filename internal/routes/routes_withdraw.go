package routes

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/vzveiteskostrami/diploma-bonus-system/internal/auth"
	"github.com/vzveiteskostrami/diploma-bonus-system/internal/dbf"
	"github.com/vzveiteskostrami/diploma-bonus-system/internal/logging"
	"github.com/vzveiteskostrami/diploma-bonus-system/internal/misc"
)

type withdrawInfo struct {
	Order *string  `json:"order,omitempty"`
	Sum   *float32 `json:"sum,omitempty"`
}

func WithdrawPostf(w http.ResponseWriter, r *http.Request) {
	var wi withdrawInfo
	if err := json.NewDecoder(r.Body).Decode(&wi); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if wi.Order == nil || wi.Sum == nil || *wi.Order == "" || *wi.Sum == 0 {
		http.Error(w, "данные неполны", http.StatusInternalServerError)
		return
	}

	if *wi.Order == "" || !misc.CheckLuhn(*wi.Order) {
		s := `Неверный номер заказа`
		http.Error(w, s, http.StatusUnprocessableEntity)
		logging.S().Infoln(s, ":", *wi.Order)
		return
	}

	userID := r.Context().Value(auth.CPuserID).(int64)
	balance, err := dbf.GetUserBalance(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logging.S().Infoln(err)
		return
	}

	if *balance.Current-*wi.Sum < 0 {
		s := "На счету недостаточно средств"
		http.Error(w, s, http.StatusPaymentRequired)
		logging.S().Infoln(s, ":", *wi.Order)
		return
	}

	code, err := dbf.WithdrawAccrual(userID, *wi.Order, *wi.Sum)
	if err != nil {
		http.Error(w, err.Error(), code)
	} else {
		w.WriteHeader(code)
	}
}

func WithdrawGetf(w http.ResponseWriter, r *http.Request) {
	completed := make(chan struct{})

	var list []dbf.Withdraw
	var err error

	go func() {
		list, err = dbf.GetUserWithdraw(r.Context().Value(auth.CPuserID).(int64))
		completed <- struct{}{}
	}()

	select {
	case <-completed:
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			if len(list) == 0 {
				w.WriteHeader(http.StatusNoContent)
			} else {
				var buf bytes.Buffer
				if err := json.NewEncoder(&buf).Encode(list); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write(buf.Bytes())
			}
		}
	case <-r.Context().Done():
		logging.S().Infoln("Получение данных прервано на клиентской стороне")
		w.WriteHeader(http.StatusGone)
	}
}
