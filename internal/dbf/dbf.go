package dbf

import (
	"net/http"

	_ "github.com/lib/pq"
)

var Store GSStorage

func MakeStorage() {
	var s PGStorage
	Store = &s
}

type GSStorage interface {
	DBFInit() int64
	DBFClose()
	DBFSaveLink(storageURLItem *StorageURL) error
	FindLink(link string, byLink bool) (StorageURL, bool)
	PingDBf(w http.ResponseWriter, r *http.Request)
	DBFGetOwnURLs(ownerID int64) ([]StorageURL, error)
	AddToDel(surl string)
	BeginDel()
	EndDel()
	PrintDBF()
}

type StorageURL struct {
	OWNERID     int64  `json:"ownerid"`
	UUID        int64  `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	Deleted     bool   `json:"deleted"`
}
