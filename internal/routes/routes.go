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

	code, err := dbf.Store.WithdrawAccrual(*wi.Order, *wi.Sum)
	if err != nil {
		http.Error(w, err.Error(), code)
	} else {
		w.WriteHeader(code)
	}
}

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
	var orders []dbf.Order
	var err error

	//logging.S().Infoln("!!!!!!", "User:", r.Context().Value(auth.CPuserID).(int64))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	orders, err = dbf.Store.GetUserOrders(r.Context().Value(auth.CPuserID).(int64))

	if err != nil {
		//logging.S().Infoln("!!!!!!", "error:", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		if len(orders) == 0 {
			//logging.S().Infoln("!!!!!!", "Пусто")
			http.Error(w, "Нет данных для ответа", http.StatusNoContent)
			w.Write([]byte("{}"))
		} else {
			var buf bytes.Buffer
			if err := json.NewEncoder(&buf).Encode(orders); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			//logging.S().Infoln("!!!!!!", "Отдали")
			w.Write(buf.Bytes())
			logging.S().Infoln("Отдали", buf)
		}
	}
}

/*
func OrdersGetf(w http.ResponseWriter, r *http.Request) {
	completed := make(chan struct{})

	var orders []dbf.Order
	var err error

	logging.S().Infoln("!!!!!!", "User:", r.Context().Value(auth.CPuserID).(int64))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	go func() {
		orders, err = dbf.Store.GetUserOrders(r.Context().Value(auth.CPuserID).(int64))
		completed <- struct{}{}
	}()

	select {
	case <-completed:
		if err != nil {
			logging.S().Infoln("!!!!!!", "errorr:", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			if len(orders) == 0 {
				logging.S().Infoln("!!!!!!", "Пусто")
				http.Error(w, "Нет данных для ответа", http.StatusNoContent)
				w.Write([]byte("{}"))
			} else {
				var buf bytes.Buffer
				if err := json.NewEncoder(&buf).Encode(orders); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				logging.S().Infoln("!!!!!!", "Отдали")
				w.Write(buf.Bytes())
			}
		}
	case <-r.Context().Done():
		logging.S().Infoln("!!!!!!", "Долго")
		logging.S().Infoln("Получение данных прервано на клиентской стороне")
		w.WriteHeader(http.StatusGone)
	}
}
*/

func BalanceGetf(w http.ResponseWriter, r *http.Request) {
	completed := make(chan struct{})

	var balance dbf.Balance
	var err error

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")

	go func() {
		balance, err = dbf.Store.GetUserBalance(r.Context().Value(auth.CPuserID).(int64))
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
			w.Write(buf.Bytes())
		}
	case <-r.Context().Done():
		logging.S().Infoln("Получение данных прервано на клиентской стороне")
		w.WriteHeader(http.StatusGone)
	}
}

func WithdrawGetf(w http.ResponseWriter, r *http.Request) {
	completed := make(chan struct{})

	var list []dbf.Withdraw
	var err error

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")

	go func() {
		list, err = dbf.Store.GetUserWithdraw(r.Context().Value(auth.CPuserID).(int64))
		completed <- struct{}{}
	}()

	select {
	case <-completed:
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			var buf bytes.Buffer
			if err := json.NewEncoder(&buf).Encode(list); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Write(buf.Bytes())
		}
	case <-r.Context().Done():
		logging.S().Infoln("Получение данных прервано на клиентской стороне")
		w.WriteHeader(http.StatusGone)
	}
}

func Registerf(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	regIn, err := misc.ExtractRegInfo(io.NopCloser(bytes.NewBuffer(body)))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	completed := make(chan struct{})

	code := http.StatusOK

	go func() {
		code, err = dbf.Store.Register(regIn.Login, regIn.Password)
		completed <- struct{}{}
	}()

	select {
	case <-completed:
		if err != nil {
			http.Error(w, err.Error(), code)
		} else {
			r.Body = io.NopCloser(bytes.NewBuffer(body))
			Authf(w, r)
		}
	case <-r.Context().Done():
		logging.S().Infow("Регистрация прервана на клиентской стороне")
		w.WriteHeader(http.StatusGone)
	}
}

func Authf(w http.ResponseWriter, r *http.Request) {
	regIn, err := misc.ExtractRegInfo(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	completed := make(chan struct{})

	code := http.StatusOK
	token := ""

	go func() {
		token, code, err = dbf.Store.Authent(regIn.Login, regIn.Password)
		completed <- struct{}{}
	}()

	select {
	case <-completed:
		if err != nil {
			http.Error(w, err.Error(), code)
		} else {
			http.SetCookie(w, &http.Cookie{Name: "token", Value: token, HttpOnly: true})
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(code)
			w.Write([]byte(token))
		}
	case <-r.Context().Done():
		logging.S().Infow("Регистрвация прервана на клиентской стороне")
		w.WriteHeader(http.StatusGone)
	}
}
