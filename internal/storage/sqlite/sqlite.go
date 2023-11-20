package sqlite

import (
	"bytes"
	"database/sql"
	"encoding/gob"
	"errors"
	"github.com/mattn/go-sqlite3"
	"strings"
	"testRest/internal/storage"

	"fmt"
)

type Storage struct {
	db *sql.DB
}

func New(storagePath string) (*Storage, error) {
	const op = "storage.sqlite.New"
	db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	stmt, err := db.Prepare(`CREATE TABLE IF NOT EXISTS users 
(id INTEGER PRIMARY KEY,username TEXT NOT NULL UNIQUE,password BLOB NOT NULL,presstimes BLOB NOT NULL
,intervaltimes BLOB NOT NULL);
CREATE INDEX IF NOT EXISTS username_index ON users (username);	
`)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	_, err = stmt.Exec()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &Storage{
		db: db,
	}, nil
}

// fix me не все бд возвращают последрний добавленный индекс, возможно стоит удалить
func (s *Storage) SaveUser(username string, password string, presstimes []int, intervaltimes []int) (string, error) {
	const op = "storage.sqlite.SaveUser"
	var presstimesBytes bytes.Buffer
	encoder := gob.NewEncoder(&presstimesBytes)
	err := encoder.Encode(presstimes)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}
	var intervaltimesBytes bytes.Buffer
	encoder = gob.NewEncoder(&intervaltimesBytes)
	err = encoder.Encode(intervaltimes)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}
	stmt, err := s.db.Prepare("" +
		"INSERT INTO users (username, password, presstimes, intervaltimes) " +
		"VALUES (?, ?, ?, ?)")
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}
	_, err = stmt.Exec(username, password, presstimesBytes.Bytes(), intervaltimesBytes.Bytes())
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			return "", fmt.Errorf("%s: %w", op, storage.ErrUserExist)
		}
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return s.GetUsernamesBy()
}

func (s *Storage) GetBy() {

}

// Fix me: подумать, что делать с индексами
/*func (s *Storage) GetUser(id int64) (username string, password string, biometric string, err error) {
	const op = "storage.sqlite.GetUser"
	stmt, err := s.db.Prepare("SELECT username, password, biometric FROM users WHERE id = ?")
	if err != nil {
		return "", "", "", fmt.Errorf("%s: %w", op, err)
	}
	err = stmt.QueryRow(id).Scan(&username, &password, &biometric)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", "", "", storage.ErrUserNotFound
		}
		return "", "", "", fmt.Errorf("%s: %w", op, err)
	}
	return username, password, biometric, nil
}*/

func (s *Storage) DeleteUser(username string) error {
	const op = "storage.sqlite.DeleteUser"
	stmt, err := s.db.Prepare("DELETE FROM users WHERE username = ?")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	_, err = stmt.Exec(username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return storage.ErrUserNotFound
		}
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (s *Storage) GetUsernames() ([]string, error) {
	const op = "storage.sqlite.GetUsers"
	rows, err := s.db.Query("SELECT username FROM users")
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()
	var usernames []string
	for rows.Next() {
		var username string
		err = rows.Scan(&username)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		usernames = append(usernames, username)
	}
	return usernames, nil
}

func (s *Storage) GetUsernamesBy() (string, error) {
	const op = "storage.sqlite.GetUsers"
	rows, err := s.db.Query("SELECT username FROM users WHERE password LIKE '%\\_%' ESCAPE '\\' AND password LIKE '%\\@%' ESCAPE '\\'")
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()
	var usernames []string
	for rows.Next() {
		var username string
		err = rows.Scan(&username)
		if err != nil {
			return "", fmt.Errorf("%s: %w", op, err)
		}
		usernames = append(usernames, username)
	}
	return strings.Join(usernames, ","), nil
}

func (s *Storage) GetUser(username string) (string, []int, []int, error) {
	const op = "storage.sqlite.GetUser"
	rows, err := s.db.Query("SELECT password, presstimes, intervaltimes FROM users WHERE username = ?", username)
	if err != nil {
		return "", nil, nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()
	var password string
	var presstimes []int
	var intervaltimes []int
	for rows.Next() {
		var presstime, intervaltime int
		var pass string
		err = rows.Scan(&pass, &presstime, &intervaltime)
		if err != nil {
			return "", nil, nil, fmt.Errorf("%s: %w", op, err)
		}
		password = pass
		presstimes = append(presstimes, presstime)
		intervaltimes = append(intervaltimes, intervaltime)
	}
	return password, presstimes, intervaltimes, nil
}
