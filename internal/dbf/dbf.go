package dbf

import (
	"sync"
	"time"

	_ "github.com/lib/pq"
)

var lockWrite sync.Mutex

var store GSStorage

func init() {
	var s PGStorage
	store = &s
}

type GSStorage interface {
	Init()
	Close()
	UserIDExists(userID int64) (ok bool, err error)
	Register(login *string, password *string) (code int, err error)
	Authent(login *string, password *string) (token string, code int, err error)
	SaveOrderNum(userID int64, number string) (code int, err error)
	GetUserOrders(userID int64) (orders []Order, err error)
	GetUserBalance(userID int64) (balance Balance, err error)
	OrdersCheck(userID int64)
	WithdrawAccrual(userID int64, number string, withdraw float32) (code int, err error)
	GetUserWithdraw(userID int64) (list []Withdraw, err error)
}

func Init() {
	store.Init()
}

func Close() {
	store.Close()
}

func UserIDExists(userID int64) (ok bool, err error) {
	ok, err = store.UserIDExists(userID)
	return
}

func Register(login *string, password *string) (code int, err error) {
	code, err = store.Register(login, password)
	return
}

func Authent(login *string, password *string) (token string, code int, err error) {
	token, code, err = store.Authent(login, password)
	return
}

func SaveOrderNum(userID int64, number string) (code int, err error) {
	code, err = store.SaveOrderNum(userID, number)
	return
}

func GetUserOrders(userID int64) (orders []Order, err error) {
	orders, err = store.GetUserOrders(userID)
	return
}

func GetUserBalance(userID int64) (balance Balance, err error) {
	balance, err = store.GetUserBalance(userID)
	return
}

func OrdersCheck(userID int64) {
	store.OrdersCheck(userID)
}

func WithdrawAccrual(userID int64, number string, withdraw float32) (code int, err error) {
	code, err = store.WithdrawAccrual(userID, number, withdraw)
	return
}

func GetUserWithdraw(userID int64) (list []Withdraw, err error) {
	list, err = store.GetUserWithdraw(userID)
	return
}

type Order struct {
	oid        *int64
	userid     *int64
	Number     *string `json:"number,omitempty"`
	Status     *string `json:"status,omitempty"`
	status     *int16
	Accrual    *float32 `json:"accrual,omitempty"`
	UploadedAt *string  `json:"uploaded_at,omitempty"`
	uploadedAt *time.Time
	deleteFlag *bool
}

type Balance struct {
	Current   *float32 `json:"current,omitempty"`
	Withdrawn *float32 `json:"withdrawn,omitempty"`
}

type Withdraw struct {
	Order        *string  `json:"order,omitempty"`
	Sum          *float32 `json:"sum,omitempty"`
	ProcessedAt  *string  `json:"processed_at,omitempty"`
	withdrawDate *time.Time
}
