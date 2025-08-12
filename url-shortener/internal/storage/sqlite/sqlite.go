package sqlite

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/hihikaAAa/GoProjects/url-shortener/internal/storage"
	sqlite3 "github.com/mattn/go-sqlite3"
)

type Storage struct {
	db *sql.DB // Коннект до базы
}

func New(storagePath string)(*Storage, error){
	const op = "storage.sqlite.New" // Константа для того, чтобы при враппинге ошибок было известно, в какой функции она произошла. Иногда добавляется в логгер

	db, err := sql.Open("sqlite3", storagePath)
	if err!= nil{
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	
	stmt,err := db.Prepare(`
	CREATE TABLE IF NOT EXISTS url(
		id INTEGER PRIMARY KEY,
		alias TEXT NOT NULL UNIQUE,
		url TEXT NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP);                                       
	CREATE TRIGGER IF NOT EXISTS url_set_updated_at
	AFTER UPDATE ON url
	FOR EACH ROW
	BEGIN
  	UPDATE url SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
	END;`)    // url - изначальная ссылка, alias - новая сокращенная, UNIQUE - 2-х записей с одинаковыми alias-ами не можеит быть
	if err!= nil{
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt.Exec()
	if err !=nil{
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &Storage{db: db},nil
}

func (s *Storage) SaveURL(urlToSave string, alias string) (int64, error){
	const op = "storage.sqlite.SaveURL"
	stmt,err := s.db.Prepare("INSERT INTO url(url,alias) VALUES(?, ?)") // Подготовить запрос
	
	if err != nil{
		return 0, fmt.Errorf("%s: %w,",op ,err)
	}
	res,err := stmt.Exec(urlToSave,alias) // Выполнить запрос
	if err != nil{
		if sqliteErr, ok := err.(sqlite3.Error); ok && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique{
			return 0, fmt.Errorf("%s: %w", op, storage.ErrUrlExists)
		}
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	id, err := res.LastInsertId() 
	if err!=nil{
		return 0, fmt.Errorf("%s : failed to get last insert id : %w", op, err)
	}
	return id, nil
}

func (s *Storage) GetURL(alias string)(string, error){
	const op = "storage.sqlite.GetURL"

	stmt, err := s.db.Prepare("SELECT url FROM url where alias=?")
	if err != nil{
		return "", fmt.Errorf("%s: %w", op, err)
	}

	var rURL string
	err = stmt.QueryRow(alias).Scan(&rURL) // Найти ряд 
	if errors.Is(err,sql.ErrNoRows){
		return "", fmt.Errorf("%s : %w", op, storage.ErrURLNotFound)
	}
	if err != nil{
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return rURL, nil
}

func (s *Storage) DeleteURL(alias string) error{
	const op = "storage.sqlite.DeleteURL"

	res,err := s.db.Exec("DELETE FROM url where alias=?") // одноразовое выполнение
	if err !=nil{
		return fmt.Errorf("%s: %w",op,err)
	}

	n, err := res.RowsAffected() // функция, показывающая сколько рядов удалили
	if err != nil{
		return fmt.Errorf("%s: get rows affected: %w", op, err)
	}
	if n == 0{
		return fmt.Errorf("%s: %w", op, storage.ErrURLNotFound)
	}
	return nil
}


