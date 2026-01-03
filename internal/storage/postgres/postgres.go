package postgres

import (
	"database/sql"
	"fmt"
	"url-shortener/internal/storage"

	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

type Storage struct {
	db *sql.DB
}

func New(connString string) (*Storage, error) {
	const op = "storage.postgres.New"

	db, err := sql.Open("postgres", connString)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	//// create the table in public schema
	//stmt := `
	//    CREATE TABLE IF NOT EXISTS public.url(
	//        id    SERIAL PRIMARY KEY,
	//        alias TEXT NOT NULL UNIQUE,
	//        url   TEXT NOT NULL
	//    );
	//`
	//fmt.Println("DEBUG: Executing CREATE TABLE...")
	//_, err = db.Exec(stmt)
	//if err != nil {
	//	fmt.Println("DEBUG: CREATE TABLE failed:", err)
	//	return nil, fmt.Errorf("%s: %w", op, err)
	//}
	//fmt.Println("DEBUG: CREATE TABLE succeeded")
	//
	//// create index in public schema
	//_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_alias ON public.url(alias);`)
	//if err != nil {
	//	return nil, fmt.Errorf("%s: %w", op, err)
	//}

	return &Storage{db: db}, nil
}

func (s *Storage) SaveURL(urlToSave string, alias string) (int64, error) {
	const op = "storage.postgres.SaveURL"

	var id int64
	err := s.db.QueryRow(
		`INSERT INTO public.url(url, alias) VALUES($1, $2) RETURNING id`,
		urlToSave, alias,
	).Scan(&id)
	if err != nil {
		// check if it's a unique constraint violation - duplicate alias
		if pqErr, ok := err.(*pq.Error); ok {
			// postgresql code 23505 means this
			if pqErr.Code == "23505" {
				return 0, fmt.Errorf("%s: %w", op, storage.ErrUrlExists)
			}
		}
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	return id, nil
}

func (s *Storage) GetURL(alias string) (string, error) {
	const op = "storage.postgres.GetURL"

	var urlToGet string
	err := s.db.QueryRow(`SELECT url FROM public.url WHERE alias=$1;`, alias).Scan(&urlToGet)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("%s: %w", op, storage.ErrURLNotFound)
		}
		return "", fmt.Errorf("%s: %w", op, err)
	}
	return urlToGet, nil
}

func (s *Storage) DeleteURL(alias string) error {
	const op = "storage.postgres.DeleteURL"

	result, err := s.db.Exec(`DELETE FROM public.url WHERE alias=$1`, alias)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	rows, err := result.RowsAffected()

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if rows == 0 {
		// no rows where deleted
		return fmt.Errorf("%s: %w", op, storage.ErrNoURLDeleted)
	}
	return nil
}
