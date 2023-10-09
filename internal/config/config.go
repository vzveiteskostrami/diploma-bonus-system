package config

import (
	"errors"
	"flag"
	"fmt"
	"os"
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
	flag.Var(Addresses.Out, "r", "Out net address host:port")
	dbc := flag.String("d", "", "Database connect string")

	flag.Parse()

	Storage.DBConnect = *dbc

	var err error
	if s, ok := os.LookupEnv("RUN_ADDRESS"); ok && s != "" {
		Addresses.In.Host, Addresses.In.Port, err = getAddrAndPort(s)
		if err != nil {
			fmt.Println("Неудачный парсинг переменной окружения RUN_ADDRESS")
		}
	}
	if s, ok := os.LookupEnv("ACCRUAL_SYSTEM_ADDRESS"); ok && s != "" {
		Addresses.Out.Host, Addresses.Out.Port, err = getAddrAndPort(s)
		if err != nil {
			fmt.Println("Неудачный парсинг переменной окружения ACCRUAL_SYSTEM_ADDRESS")
		}
	}
	if s, ok := os.LookupEnv("DATABASE_URI"); ok && s != "" {
		Storage.DBConnect = s
	}

	// сохранена/закомментирована эмуляция указания БД в параметрах вызова.
	// Необходимо для быстрого перехода тестирования работы приложения с
	// Postgres.
	//Storage.DBConnect = "host=127.0.0.1 port=5432 user=executor password=executor dbname=gophermart sslmode=disable"
}
