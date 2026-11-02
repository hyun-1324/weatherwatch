package watchlist

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

const selectColumns = `SELECT id, external_id, name, country, admin1, latitude, longitude, created_at FROM locations`

func (s *Store) List(ctx context.Context) ([]Location, error) {
	rows, err := s.db.QueryContext(ctx, selectColumns+` ORDER BY created_at, id`)
	if err != nil {
		return nil, fmt.Errorf("list locations: %w", err)
	}
	defer rows.Close()

	locations := []Location{}
	for rows.Next() {
		loc, err := scanLocation(rows)
		if err != nil {
			return nil, fmt.Errorf("scan location: %w", err)
		}
		locations = append(locations, loc)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate locations: %w", err)
	}
	return locations, nil
}

func (s *Store) Get(ctx context.Context, id int64) (Location, error) {
	loc, err := scanLocation(s.db.QueryRowContext(ctx, selectColumns+` WHERE id = ?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return Location{}, ErrLocationNotFound
	}
	if err != nil {
		return Location{}, fmt.Errorf("get location %d: %w", id, err)
	}
	return loc, nil
}

func (s *Store) Create(ctx context.Context, loc Location) (Location, error) {
	loc.CreatedAt = time.Now().UTC().Truncate(time.Second)
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO locations (external_id, name, country, admin1, latitude, longitude, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		loc.ExternalID, loc.Name, loc.Country, loc.Admin1, loc.Latitude, loc.Longitude,
		loc.CreatedAt.Format(time.RFC3339),
	)
	if err != nil {
		if sqliteErr, ok := errors.AsType[*sqlite.Error](err); ok && sqliteErr.Code() == sqlite3.SQLITE_CONSTRAINT_UNIQUE {
			return Location{}, ErrLocationAlreadySaved
		}
		return Location{}, fmt.Errorf("insert location: %w", err)
	}
	if loc.ID, err = res.LastInsertId(); err != nil {
		return Location{}, fmt.Errorf("read inserted id: %w", err)
	}
	return loc, nil
}

func (s *Store) Delete(ctx context.Context, id int64) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM locations WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete location %d: %w", id, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete location %d: %w", id, err)
	}
	if n == 0 {
		return ErrLocationNotFound
	}
	return nil
}

type scanner interface{ Scan(dest ...any) error }

func scanLocation(row scanner) (Location, error) {
	var (
		loc       Location
		createdAt string
	)
	err := row.Scan(&loc.ID, &loc.ExternalID, &loc.Name, &loc.Country, &loc.Admin1,
		&loc.Latitude, &loc.Longitude, &createdAt)
	if err != nil {
		return Location{}, err
	}
	if loc.CreatedAt, err = time.Parse(time.RFC3339, createdAt); err != nil {
		return Location{}, fmt.Errorf("parse created_at %q: %w", createdAt, err)
	}
	return loc, nil
}
