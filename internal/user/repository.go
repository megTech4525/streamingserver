package user

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
)

type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateUser(ctx context.Context, u *User) error {
	// NamedQueryContext lets you use :field names that match the db tags,
	// so you don't have to line up $1, $2, $3 by hand.
	rows, err := r.db.NamedQueryContext(ctx,
		`INSERT INTO users (username, email, password)
		 VALUES (:username, :email, :password)
		 RETURNING id, created_at`, u)
	if err != nil {
		return err
	}
	defer rows.Close()

	if rows.Next() {
		return rows.Scan(&u.ID, &u.CreatedAt)
	}
	return rows.Err()
}

func (r *Repository) GetByEmail(ctx context.Context, email string) (*User, error) {
	var u User
	err := r.db.GetContext(ctx, &u,
		`SELECT id, username, email, password, created_at
		 FROM users WHERE email = $1`, email)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}
