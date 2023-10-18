package checkout

import (
	"net/http"

	"github.com/vzveiteskostrami/diploma-bonus-system/internal/auth"
	"github.com/vzveiteskostrami/diploma-bonus-system/internal/dbf"
)

func RunCheck(h http.Handler) http.Handler {
	he := func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(auth.CPuserID).(int64)
		go func() {
			dbf.OrdersCheck(userID)
		}()
		h.ServeHTTP(w, r)
	}
	return http.HandlerFunc(he)
}
