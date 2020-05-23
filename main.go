package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"log"
	"net/http"

	"golang.org/x/sync/errgroup"

	_ "github.com/go-sql-driver/mysql"
)

var addr = flag.String("addr", ":8812", "Listen for HTTP at this address")
var dbconn = flag.String("dbconn", "root:rootpw@tcp4(127.0.0.1:3306)/testm?collation=utf8mb4_unicode_ci", "Connect to mysql with this connection string")

func main() {
	flag.Parse()
	log.Printf("Starting testm, listener at %q with conn %q", *addr, *dbconn)

	var db *sql.DB

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		k := r.URL.Path

		if db == nil {
			http.Error(w, "database not available", 500)
			return
		}

		_, err := db.Exec(`INSERT INTO testm_counter(k,c) VALUES(?,0) ON DUPLICATE KEY UPDATE c = c + 1`, k)
		if err != nil {
			panic(err)
		}

		rows, err := db.Query(`SELECT c FROM testm_counter WHERE k = ?`, k)
		if err != nil {
			panic(err)
		}
		defer rows.Close()
		var c int
		for rows.Next() {
			rows.Scan(&c)
		}

		log.Printf("URL: %s -> %d", k, c)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"c": c})
	})

	var g errgroup.Group

	g.Go(func() error { return http.ListenAndServe(*addr, h) })

	g.Go(func() (reterr error) {
		defer func() {
			if reterr != nil {
				log.Printf("MySQL setup error: %v", reterr)
			}
		}()

		var err error
		db, err = sql.Open("mysql", *dbconn)
		if err != nil {
			return err
		}
		defer db.Close()

		_, err = db.Exec(`CREATE TABLE IF NOT EXISTS testm_counter (
			k VARCHAR(255) NOT NULL,
			c INT NOT NULL,
			PRIMARY KEY (k)
		)`)
		if err != nil {
			return err
		}
		return nil
	})

	err := g.Wait()
	if err != nil {
		log.Fatal(err)
	}

}
