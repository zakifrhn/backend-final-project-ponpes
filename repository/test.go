package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

// Test represents a simple entity stored in the repository.
type Test struct {
	ID        int64     `db:"id"`
	Name      string    `db:"name"`
	CreatedAt time.Time `db:"created_at"`
}

// Common repository errors.
var ErrNotFound = errors.New("repository: not found")

// TestRepository defines basic CRUD operations for Test entity.
type TestRepository interface {
	Create(ctx context.Context, t *Test) error
	GetByID(ctx context.Context, id int64) (*Test, error)
	Update(ctx context.Context, t *Test) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, limit, offset int) ([]*Test, error)
}

// sqlTestRepo is a SQL-backed implementation of TestRepository.
type sqlTestRepo struct {
	db *sql.DB
}

// NewTestRepository creates a new SQL-backed TestRepository.
// Note: SQL placeholder syntax (?) may vary by driver (use $1... for Postgres).
func NewTestRepository(db *sql.DB) TestRepository {
	return &sqlTestRepo{db: db}
}

func (r *sqlTestRepo) Create(ctx context.Context, t *Test) error {
	if t == nil {
		return errors.New("repository: test is nil")
	}
	if t.CreatedAt.IsZero() {
		t.CreatedAt = time.Now().UTC()
	}

	query := `INSERT INTO tests (name, created_at) VALUES (?, ?)`
	res, err := r.db.ExecContext(ctx, query, t.Name, t.CreatedAt)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	t.ID = id
	return nil
}

func (r *sqlTestRepo) GetByID(ctx context.Context, id int64) (*Test, error) {
	query := `SELECT id, name, created_at FROM tests WHERE id = ?`
	var t Test
	err := r.db.QueryRowContext(ctx, query, id).Scan(&t.ID, &t.Name, &t.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *sqlTestRepo) Update(ctx context.Context, t *Test) error {
	if t == nil {
		return errors.New("repository: test is nil")
	}
	query := `UPDATE tests SET name = ? WHERE id = ?`
	res, err := r.db.ExecContext(ctx, query, t.Name, t.ID)
	if err != nil {
		return err
	}
	ra, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if ra == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *sqlTestRepo) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM tests WHERE id = ?`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	ra, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if ra == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *sqlTestRepo) List(ctx context.Context, limit, offset int) ([]*Test, error) {
	query := `SELECT id, name, created_at FROM tests ORDER BY id DESC LIMIT ? OFFSET ?`
	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*Test
	for rows.Next() {
		var t Test
		if err := rows.Scan(&t.ID, &t.Name, &t.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, &t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return list, nil
}
