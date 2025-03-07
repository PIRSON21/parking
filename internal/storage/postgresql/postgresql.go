package postgresql

import (
	"database/sql"
	"fmt"
	"github.com/PIRSON21/parking/internal/config"
	"log"

	_ "github.com/lib/pq"
)

// MustConnectDB подключает к базе данных PostgreSQL по данным конфига:
// username, dbname, password. Остальные параметры стандартные
func MustConnectDB(cfg *config.Config) *sql.DB {
	connStr := fmt.Sprintf(
		"user='%s' dbname='%s' password='%s' sslmode=disable",
		cfg.DBUsername, cfg.DBName, cfg.DBPassword,
	)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("error while connecting to DB: ", err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal("error when ping to DB server: ", err)
	}

	return db
}
