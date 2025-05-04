package postgresql

import (
	"database/sql"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/pashagolub/pgxmock/v4"
	"testing"
)

func TestStorage_fetchParkings(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error while creating mocks: %v", err)
	}

	s := &Storage{db}

	query := `SELECT 
		    parking_id, parking_name, parking_address, parking_width, parking_height, day_tariff, night_tariff, parking_topology
		FROM parkings
		WHERE parking_name ILIKE $1 AND manager_id = $2`
	mock.ExpectPrepare(query)
	mock.ExpectQuery(query).WillReturnRows(sqlmock.NewRows())
}
