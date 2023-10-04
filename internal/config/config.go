package config

import (
	"errors"
	"flag"
	"strconv"
	"strings"
)

var (
	Addresses InOutAddresses
	Storage   StorageAttr
)

type NetAddress struct {
	Host string
	Port int
}

func (na *NetAddress) String() string {
	return na.Host + ":" + strconv.Itoa(na.Port)
}

func (na *NetAddress) Set(flagValue string) error {
	var err error
	na.Host, na.Port, err = getAddrAndPort(flagValue)
	return err
}

type StorageAttr struct {
	DBConnect string
}

func getAddrAndPort(s string) (string, int, error) {
	var err error
	h := ""
	p := int(0)
	args := strings.Split(s, ":")
	if len(args) == 2 || len(args) == 3 {
		if len(args) == 3 {
			h = args[0] + ":" + args[1]
			args[1] = args[2]
		} else {
			h = args[0]
		}

		if args[1] == "" {
			return "", -1, errors.New("неверный формат строки, требуется host:port")
		}
		p, err = strconv.Atoi(args[1])
		if err != nil {
			return "", -1, errors.New("неверный номер порта, " + err.Error())
		}
	} else {
		return "", -1, errors.New("неверный формат строки, требуется host:port")
	}
	return h, p, nil
}

type InOutAddresses struct {
	In  *NetAddress
	Out *NetAddress
}

func ReadData() {
	Addresses.In = new(NetAddress)
	Addresses.Out = new(NetAddress)
	Addresses.In.Host = "localhost"
	Addresses.In.Port = 8080
	Addresses.Out.Host = "http://127.0.0.1"
	Addresses.Out.Port = 8080

	_ = flag.Value(Addresses.In)
	flag.Var(Addresses.In, "a", "In net address host:port")
	_ = flag.Value(Addresses.Out)
	flag.Var(Addresses.Out, "b", "Out net address host:port")
	dbc := flag.String("d", "", "Database connect string")

	flag.Parse()

	Storage.DBConnect = *dbc
	// сохранена/закомментирована эмуляция указания БД в параметрах вызова.
	// Необходимо для быстрого перехода тестирования работы приложения с
	// Postgres.
	Storage.DBConnect = "host=127.0.0.1 port=5432 user=executor password=executor dbname=gophermart sslmode=disable"
}
