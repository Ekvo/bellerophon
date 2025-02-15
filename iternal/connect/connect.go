package connect

import (
	"encoding/json"
	"fmt"
	"os"
)

type Connect struct {
	Host         string `json:"host"`
	Port         string `json:"port"`
	User         string `json:"user"`
	Password     string `json:"password"`
	DataBaseName string `json:"database"`
	Sslmode      string `json:"sslmode"`
}

func (c Connect) String() string {
	return fmt.Sprintf(`
host=%s port=%s
user=%s password=%s dbname=%s
sslmode=%s`,
		c.Host, c.Port,
		c.User, c.Password, c.DataBaseName,
		c.Sslmode)
}

func NewConnect(fileName string) (*Connect, error) {
	connectData, errData := os.Open(fileName)
	if errData != nil {
		return nil, errData
	}
	dec := json.NewDecoder(connectData)
	dec.DisallowUnknownFields()

	var conn Connect
	errDec := dec.Decode(&conn)
	if errDec != nil {
		return nil, errDec
	}

	return &conn, nil
}
