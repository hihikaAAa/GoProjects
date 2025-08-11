package sqlite

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3" 
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