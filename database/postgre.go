package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

var PostgreDB *sql.DB

func ConnectPostgre() {
	var err error

	// Sesuaikan dengan environment sistem Anda
	dsn := "host=localhost user=postgres password=@cindy1501 dbname=prestasi_db port=5432 sslmode=disable"

	PostgreDB, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("Gagal koneksi ke PostgreSQL:", err)
	}

	if err = PostgreDB.Ping(); err != nil {
		log.Fatal("Gagal ping PostgreSQL:", err)
	}

	fmt.Println("Berhasil terhubung ke PostgreSQL")
}
