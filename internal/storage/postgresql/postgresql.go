package postgresql

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/PIRSON21/parking/internal/config"
	"github.com/PIRSON21/parking/internal/models"
	"log"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// Storage хранит закрытое в пакете соединение с БД
type Storage struct {
	db *sql.DB
}

// MustConnectDB подключает к базе данных PostgresSQL по данным конфига: username, dbname, password.
// Остальные параметры стандартные
func MustConnectDB(cfg *config.Config) *Storage {
	connStr := fmt.Sprintf(
		"user='%s' dbname='%s' password='%s' sslmode=disable",
		cfg.DBUsername, cfg.DBName, cfg.DBPassword,
	)
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		log.Fatal("error while connecting to DB: ", err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal("error when ping to DB server: ", err)
	}

	return &Storage{db}
}

// GetParkingsList обращается в БД и получает список всех парковок (+ поиск по имени)
func (s *Storage) GetParkingsList(search string) ([]*models.Parking, error) {
	const op = "storage.postgresql.GetParkingsList"

	stmt, err := s.db.Prepare(`
			SELECT 
			    parking_id, parking_name, parking_address, parking_width, parking_height 
			FROM parkings
			WHERE parking_name ILIKE $1 ;
    `)
	if err != nil {
		return nil, fmt.Errorf("%s: error while prepare statement: %w", op, err)
	}

	rows, err := stmt.Query("%" + search + "%")
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("%s: error while getting result: %w", op, err)
	}
	defer rows.Close()

	var resParking []*models.Parking

	for rows.Next() {
		var parking models.Parking

		err = rows.Scan(&parking.ID, &parking.Name, &parking.Address, &parking.Width, &parking.Height)
		if err != nil {
			log.Printf("%s: error while reading rows: %v", op, err)
		}

		resParking = append(resParking, &parking)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: error while scanning rows: %w", op, err)
	}

	return resParking, nil
}

func (s *Storage) AddParking(parking *models.Parking) error {
	const op = "storage.postgresql.AddParking"

	stmt, err := s.db.Prepare(`
		INSERT INTO parkings (parking_name, parking_address, parking_width, parking_height)
		VALUES ($1, $2, $3, $4);
	`)
	if err != nil {
		return fmt.Errorf("%s: error while preparing statement: %w", op, err)
	}

	_, err = stmt.Exec(parking.Name, parking.Address, parking.Width, parking.Height)
	if err != nil {
		return fmt.Errorf("%s: error while executing statement: %w", op, err)
	}

	return nil
}
