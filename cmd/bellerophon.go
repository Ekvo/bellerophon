package main

import (
	"database/sql"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"time"

	"github.com/Ekvo/bellerophon/iternal/app"
	"github.com/Ekvo/bellerophon/iternal/connect"
	"github.com/Ekvo/bellerophon/iternal/source"
)

func main() {
	conn, errCon := connect.NewConnect("./iternal/connect/connectData.json")
	if errCon != nil {
		log.Fatalf("no connect data - %v", errCon)
	}

	db, errDB := sql.Open("postgres", conn.String())
	if errDB != nil {
		log.Fatalf("no open DB - %v", errDB)
	}
	defer func() {
		err := db.Close()
		if err != nil {
			log.Printf("sql.DB Close - %v", err)
		}
	}()

	s := source.NewSqlSource(db)
	a := app.NewApplication(s)
	r := mux.NewRouter()

	a.Routes(r)

	srv := http.Server{
		Addr:         "127.0.0.1:8000",
		Handler:      r,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("start server error - %v", err)
	}
}
