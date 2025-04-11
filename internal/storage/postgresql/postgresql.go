package postgresql

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/PIRSON21/parking/internal/config"
	"github.com/PIRSON21/parking/internal/http-server/handler/user"
	"github.com/PIRSON21/parking/internal/lib/api/request"
	custErr "github.com/PIRSON21/parking/internal/lib/errors"
	"github.com/PIRSON21/parking/internal/models"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
	"log"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// Storage хранит закрытое в пакете соединение с БД.
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

// GetAdminParkings получает из БД данные о всех парковках по запросу search.
func (s *Storage) GetAdminParkings(search string) ([]*models.Parking, error) {
	query := `
			SELECT 
			    parking_id, parking_name, parking_address, parking_width, parking_height, day_tariff, night_tariff 
			FROM parkings
			WHERE parking_name ILIKE $1
    `
	search = "%" + search + "%"
	return s.fetchParkings(query, search)
}

// GetManagerParkings получает из БД данные о всех доступных менеджеру парковках по запросу search.
func (s *Storage) GetManagerParkings(userID int, search string) ([]*models.Parking, error) {
	query := `
			SELECT 
			    parking_id, parking_name, parking_address, parking_width, parking_height, day_tariff, night_tariff
			FROM parkings
			WHERE parking_name ILIKE $1 AND manager_id = $2
    `
	search = "%" + search + "%"

	return s.fetchParkings(query, search, userID)
}

// fetchParkings исполняет запросы и структурирует данные.
func (s *Storage) fetchParkings(query string, args ...interface{}) ([]*models.Parking, error) {
	const op = "storage.postgresql.fetchParkings"

	stmt, err := s.db.Prepare(query)
	if err != nil {
		return nil, fmt.Errorf("%s: error while prepare statement: %w", op, err)
	}

	rows, err := stmt.Query(args...)
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

		err = rows.Scan(&parking.ID, &parking.Name, &parking.Address, &parking.Width, &parking.Height, &parking.DayTariff, &parking.NightTariff)
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

// AddParking добавляет данные о парковке в БД.
func (s *Storage) AddParking(parking *models.Parking) error {
	const op = "storage.postgresql.AddParking"

	stmt, err := s.db.Prepare(`
		INSERT INTO parkings (parking_name, parking_address, parking_width, parking_height, day_tariff, night_tariff)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING parking_id;
	`)
	if err != nil {
		return fmt.Errorf("%s: error while preparing statement: %w", op, err)
	}

	err = stmt.QueryRow(parking.Name, parking.Address, parking.Width, parking.Height, parking.DayTariff, parking.NightTariff).Scan(&parking.ID)
	if err != nil {
		return fmt.Errorf("%s: error while executing statement: %w", op, err)
	}

	return nil
}

// GetParkingByID получает всю информацию (что хранится в таблице парковки) о парковке из БД..
//
// Возвращает указатель на модель парковки или ошибку.
func (s *Storage) GetParkingByID(parkingID int) (*models.Parking, error) {
	const op = "storage.postgresql.GetParkingById"

	stmt, err := s.db.Prepare(`
	SELECT 
	    parking_id, parking_name, parking_address, parking_width, parking_height, manager_id, day_tariff, night_tariff 
	FROM parkings
	WHERE parking_id = $1;
`)
	if err != nil {
		return nil, fmt.Errorf("%s: error while preparing statement: %w", op, err)
	}

	var parking models.Parking
	var managerID sql.NullInt64
	if err = stmt.QueryRow(parkingID).Scan(&parking.ID, &parking.Name, &parking.Address, &parking.Width, &parking.Height, &managerID, &parking.DayTariff, &parking.NightTariff); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}

		return nil, fmt.Errorf("%s: error while executing statement: %w", op, err)
	}

	if managerID.Valid {
		parking.Manager = &models.Manager{ID: int(managerID.Int64)}
	}

	return &parking, nil
}

// GetParkingCells получает данные о клетках из базы данных и на основе строит матрицу топологии парковки
//
//goland:noinspection t
func (s *Storage) GetParkingCells(parking *models.Parking) ([][]models.ParkingCell, error) {
	op := "storage.postgresql.GetParkingCells"

	width := parking.Width
	height := parking.Height

	parkingCells := make([][]models.ParkingCell, width)
	for x := range parkingCells {
		parkingCells[x] = make([]models.ParkingCell, height)
		for y := range parkingCells[x] {
			parkingCells[x][y] = "."
		}
	}

	stmt, err := s.db.Prepare(`
		SELECT 
			x,y, cell_type
		FROM parking_cell
		WHERE parking_id = $1;
	`)
	if err != nil {
		return nil, fmt.Errorf("%s: error while preparing statement: %w", op, err)
	}

	rows, err := stmt.Query(parking.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("%s: error while getting result from DB: %w", op, err)
	}

	defer rows.Close()

	found := false

	for rows.Next() {
		found = true
		var x, y int

		var cellType models.ParkingCell

		if err = rows.Scan(&x, &y, &cellType); err != nil {
			return nil, fmt.Errorf("%s: error while scanning result to var: %w", op, err)
		}

		fmt.Println(cellType)

		if x < 0 || x >= height || y < 0 || y >= width {
			continue
		}

		parkingCells[y][x] = cellType
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: error after scanning rows: %w", op, err)
	}

	if !found {
		return nil, nil
	}

	return parkingCells, nil

}

// AddCellsForParking добавляет информацию о клетках топологии парковки.
func (s *Storage) AddCellsForParking(parking *models.Parking, cells []*models.ParkingCellStruct) error {
	const op = "storage.postgresql.AddCellsForParking"

	if len(cells) == 0 {
		return nil
	}

	valueStrings := make([]string, 0, len(cells))
	valueArgs := make([]interface{}, 0, len(cells)*3)

	for i, cell := range cells {
		valueStrings = append(valueStrings, fmt.Sprintf("(%d, $%d, $%d, $%d)", parking.ID, i*3+1, i*3+2, i*3+3))
		valueArgs = append(valueArgs, cell.X, cell.Y, cell.CellType)
	}

	query := fmt.Sprintf(
		"INSERT INTO parking_cell (parking_id, x, y, cell_type) VALUES %s ON CONFLICT (parking_id, x, y) DO UPDATE SET cell_type = EXCLUDED.cell_type",
		strings.Join(valueStrings, ", "),
	)

	fmt.Println(query)
	fmt.Println(valueArgs)

	_, err := s.db.Exec(query, valueArgs...)
	if err != nil {
		return fmt.Errorf("%s: error while executing query: %w", op, err)
	}

	return nil
}

// GetUserID получает и проверяет актуальность сессии, и возвращает id пользователя.
// Если id > 0, то пользователь - менеджер. Если id = 0, пользователь - администратор.
// Если сессия истекла или не найдена, выведется ошибка ErrUnauthorized.
func (s *Storage) GetUserID(sessionID string) (int, error) {
	const op = "storage.postgresql.GetUserID"

	stmt, err := s.db.Prepare(`
	SELECT user_id, deadline 
	FROM user_session 
	WHERE session_id = $1;
	`)
	if err != nil {
		return 0, fmt.Errorf("%s: error while preparing statement: %w", op, err)
	}

	var userID sql.NullInt64
	var deadline time.Time

	err = stmt.QueryRow(sessionID).Scan(&userID, &deadline)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, custErr.ErrUnauthorized
		}

		return 0, fmt.Errorf("%s: error while getting row: %w", op, err)
	}

	if time.Now().After(deadline) {
		return 0, custErr.ErrSessionExpired
	}

	if userID.Valid {
		return int(userID.Int64), nil
	} else {
		return 0, nil
	}
}

// AuthenticateManager проверяет введенный логин и пароль на достоверность.
func (s *Storage) AuthenticateManager(user *models.User) (int, error) {
	const op = "storage.postgresql.AuthenticateManager"

	var hashedPassword string
	var managerID int

	stmt, err := s.db.Prepare(`SELECT manager_id, manager_password FROM manager WHERE manager_login = $1;`)
	if err != nil {

		return 0, fmt.Errorf("%s: error while preparing statement: %w", op, err)
	}

	err = stmt.QueryRow(user.Login).Scan(&managerID, &hashedPassword)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, custErr.ErrUnauthorized
		}
		return 0, fmt.Errorf("%s: error while executing statement: %w", op, err)
	}

	if !checkPassword(user.Password, hashedPassword) {
		return 0, custErr.ErrUnauthorized
	}

	return managerID, nil
}

func checkPassword(inputPassword, hashedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(inputPassword))
	return err == nil
}

// SetSessionID добавляет в БД информацию о сессии.
func (s *Storage) SetSessionID(userID int, sessionID string) error {
	const op = "storage.postgresql.SetSessionID"
	var queryID sql.NullInt64

	if userID != 0 {
		queryID.Int64 = int64(userID)
		queryID.Valid = true
	}

	fmt.Println(queryID)

	stmt, err := s.db.Prepare(`
	INSERT INTO user_session(session_id, user_id, deadline)
	VALUES ($1, $2, $3);
	`)
	if err != nil {
		return fmt.Errorf("%s: error while preparing statement: %w", op, err)
	}

	deadline := time.Now().Add(72 * time.Hour)

	_, err = stmt.Exec(sessionID, queryID, deadline)
	if err != nil {
		return fmt.Errorf("%s: error while executing statement: %w", op, err)
	}

	return nil
}

// CreateNewManager создает нового менеджера в БД.
// Возвращает только ошибку.
func (s *Storage) CreateNewManager(manager *request.UserCreate) error {
	const op = "storage.postgresql.CreateNewManager"

	stmt, err := s.db.Prepare(`
	INSERT INTO manager(manager_login, manager_password, manager_email)
	VALUES ($1, $2, $3);
	`)
	if err != nil {
		return fmt.Errorf("%s: error while preparing statement: %w", op, err)
	}

	hashedPassword, err := createPasswordHash(manager.Password)
	if err != nil {
		return fmt.Errorf("%s: error while creating password hash: %w", op, err)
	}

	_, err = stmt.Exec(manager.Login, hashedPassword, manager.Email)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return pgErr
			}
		}

		return fmt.Errorf("%s: error while executing statement: %w", op, err)
	}

	return nil
}

func createPasswordHash(password string) (string, error) {
	pass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(pass), nil
}

func (s *Storage) GetManagers() ([]*models.User, error) {
	const op = "storage.postgresql.GetManagers"

	stmt, err := s.db.Prepare(`
	SELECT manager_id, manager_login, manager_email
	FROM manager `)
	if err != nil {
		return nil, fmt.Errorf("%s: error while preparing statement: %w", op, err)
	}

	rows, err := stmt.Query()
	defer rows.Close()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("%s: error while getting rows: %w", op, err)
	}

	var managers []*models.User

	for rows.Next() {
		manager := new(models.User)
		err = rows.Scan(&manager.ID, &manager.Login, &manager.Email)
		if err != nil {
			return nil, fmt.Errorf("%s: error while reading rows: %w", op, err)
		}
		managers = append(managers, manager)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("%s: error with rows: %w", op, err)
	}

	return managers, nil
}

func (s *Storage) GetManagerByID(managerID int) (*models.User, error) {
	const op = "storage.postgresql.GetManagerByID"

	stmt, err := s.db.Prepare(`
	SELECT manager_id, manager_login, manager_email
	FROM manager
	WHERE manager_id = $1
	`)
	if err != nil {
		return nil, fmt.Errorf("%s: error while preparing statement: %w", op, err)
	}

	var manager models.User

	if err := stmt.QueryRow(managerID).Scan(&manager.ID, &manager.Login, &manager.Email); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("%s: error while reading rows: %w", op, err)
	}

	return &manager, nil
}

// UpdateManager обновляет данные о менеджере в БД.
func (s *Storage) UpdateManager(manager *user.UserPatch) error {
	const op = "storage.postgresql.UpdateManager"

	query := "UPDATE manager SET "
	var updates []string
	var args []interface{}
	argIdx := 1

	if manager.Login != nil {
		updates = append(updates, fmt.Sprintf("manager_login = $%d", argIdx))
		args = append(args, manager.Login)
		argIdx++
	}

	if manager.Email != nil {
		updates = append(updates, fmt.Sprintf("manager_email = $%d", argIdx))
		args = append(args, manager.Email)
		argIdx++
	}

	if manager.Password != nil {
		updates = append(updates, fmt.Sprintf("manager_password = $%d", argIdx))
		hashedPassword, err := createPasswordHash(*manager.Password)
		if err != nil {
			return fmt.Errorf("%s: error while hashing password: %w", op, err)
		}
		args = append(args, hashedPassword)
		argIdx++
	}

	query += strings.Join(updates, ", ") + fmt.Sprintf(" WHERE manager_id = $%d", argIdx)
	args = append(args, manager.ID)
	fmt.Println(query, args)

	_, err := s.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("%s: error while executing statement: %w", op, err)
	}

	return nil
}

// DeleteManager удаляет менеджера из БД.
func (s *Storage) DeleteManager(managerID int) error {
	const op = "storage.postgresql.DeleteManager"

	stmt, err := s.db.Prepare(`DELETE FROM manager WHERE manager_id = $1`)
	if err != nil {
		return fmt.Errorf("%s: error while preparing statement: %w", op, err)
	}

	_, err = stmt.Exec(managerID)
	if err != nil {
		return fmt.Errorf("%s: error while executing statement: %w", op, err)
	}

	stmt, err = s.db.Prepare(`DELETE FROM user_session WHERE user_id = $1`)
	if err != nil {
		return fmt.Errorf("%s: error while preparing statement: %w", op, err)
	}

	_, err = stmt.Exec(managerID)
	if err != nil {
		return fmt.Errorf("%s: error while executing statement: %w", op, err)
	}

	return nil
}
