package routes

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/vzveiteskostrami/diploma-bonus-system/internal/auth"
	"github.com/vzveiteskostrami/diploma-bonus-system/internal/dbf"
	"github.com/vzveiteskostrami/diploma-bonus-system/internal/logging"
)

func BalanceGetf(w http.ResponseWriter, r *http.Request) {
	completed := make(chan struct{})

	var balance dbf.Balance
	var err error

	go func() {
		balance, err = dbf.GetUserBalance(r.Context().Value(auth.CPuserID).(int64))
		completed <- struct{}{}
	}()

	select {
	case <-completed:
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			var buf bytes.Buffer
			if err := json.NewEncoder(&buf).Encode(balance); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(buf.Bytes())
		}
	case <-r.Context().Done():
		logging.S().Infoln("Получение данных прервано на клиентской стороне")
		w.WriteHeader(http.StatusGone)
	}
}
