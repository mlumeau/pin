package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "modernc.org/sqlite"
	"pin/internal/config"
	"pin/internal/features/identity"
	platformhttp "pin/internal/platform/http"
	platformserver "pin/internal/platform/server"
	"pin/internal/platform/storage"
	sqlitestore "pin/internal/platform/storage/sqlite"
)

// main wires dependencies and starts the HTTP server.
func main() {
	if len(os.Args) > 1 && os.Args[1] == "backup" {
		cfg, err := config.LoadConfig()
		if err != nil {
			log.Fatalf("config error: %v", err)
		}
		db, err := sql.Open("sqlite", cfg.DBPath)
		if err != nil {
			log.Fatalf("db open: %v", err)
		}
		defer db.Close()
		dest := ""
		if len(os.Args) > 2 {
			dest = os.Args[2]
		}
		if path, err := storage.BackupToZip(cfg, dest); err != nil {
			log.Fatalf("backup failed: %v", err)
		} else {
			log.Printf("backup written to %s", path)
		}
		return
	}
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	db, err := sql.Open("sqlite", cfg.DBPath)
	if err != nil {
		log.Fatalf("db open: %v", err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	if _, err := db.Exec("PRAGMA busy_timeout = 5000"); err != nil {
		log.Fatalf("db busy_timeout: %v", err)
	}
	if _, err := db.Exec("PRAGMA journal_mode = WAL"); err != nil {
		log.Fatalf("db journal_mode: %v", err)
	}

	if err := sqlitestore.InitDB(db); err != nil {
		log.Fatalf("db init: %v", err)
	}

	srv, err := platformserver.NewServer(cfg, db, identity.TemplateFuncs())
	if err != nil {
		log.Fatalf("server init: %v", err)
	}

	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	log.Printf("listening on %s", addr)
	if err := http.ListenAndServe(addr, platformhttp.Routes(srv)); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
