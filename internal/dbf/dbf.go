package dbf

import (
	"net/http"
	"sync"
	"time"

	_ "github.com/lib/pq"
)

var lockWrite sync.Mutex

var Store GSStorage

func MakeStorage() {
	var s PGStorage
	Store = &s
}

type GSStorage interface {
	DBFInit() int64
	DBFClose()
	PingDBf(w http.ResponseWriter, r *http.Request)
	AddToDel(surl string)
	BeginDel()
	EndDel()
	PrintDBF()
	UserIDExists(userID int64) (ok bool, err error)
	Register(login *string, password *string) (code int, err error)
	Authent(login *string, password *string) (token string, code int, err error)
	SaveOrderNum(userID int64, number string) (code int, err error)
	GetUserOrders(userID int64) (orders []Order, err error)
	GetUserBalance(userID int64) (balance Balance, err error)
}

type Order struct {
	oid        *int64
	userid     *int64
	Number     *string  `json:"number,omitempty"`
	Status     *string  `json:"status,omitempty"`
	Accrual    *float32 `json:"accural,omitempty"`
	UploadedAt *string  `json:"uploaded_at,omitempty"`
	uploadedAt *time.Time
	deleteFlag *bool
}

type Balance struct {
	Current   *float32 `json:"current,omitempty"`
	Withdrawn *float32 `json:"withdrawn,omitempty"`
}
