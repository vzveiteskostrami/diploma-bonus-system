package auth

import (
	"context"
	"errors"
	"net/http"

	"github.com/vzveiteskostrami/diploma-bonus-system/internal/dbf"
	"github.com/vzveiteskostrami/diploma-bonus-system/internal/logging"
	"github.com/vzveiteskostrami/diploma-bonus-system/internal/misc"
)

// 0 = eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJPd25lcklEIjowfQ.u6d3Bcz7A-MulX5WbdBJypc56uRF2DOILD_WxqOsvOk
// 1 = eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJPd25lcklEIjoxfQ.cOg2cbX9qBBQUH1yqvNIgMWX-w-PnXdPxr5tbmXg4fw

type ContextParamName string

var (
	CPuserID ContextParamName = "UserID"
)

func AuthHandle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var userID int64 = 0
		var ok bool

		cu, err := r.Cookie("token")

		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
		} else if userID, ok = misc.GetUserData(cu.Value); !ok {
			http.Error(w, err.Error(), http.StatusUnauthorized)
		}

		if err != nil {
			logging.S().Error(err)
			return
		}

		ok, err = dbf.Store.UserIDExists(userID)
		if err != nil {
			logging.S().Error(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		} else if !ok {
			err = errors.New("userId не найден в системе")
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		c := context.WithValue(r.Context(), CPuserID, userID)

		next.ServeHTTP(w, r.WithContext(c))
	})
}
