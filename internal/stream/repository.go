package stream

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
)

type Repo struct {
	db *sqlx.DB
}

func NewRepo(db *sqlx.DB) *Repo {
	return &Repo{db: db}
}

func (r *Repo) Create(ctx context.Context, s *Stream) error {
	return r.db.GetContext(ctx, &s.ID,
		`INSERT INTO streams (user_id, title,
		 stream_key, playback_id)
	VALUES ($1, $2, $3, $4)
	RETURNING id`,
		s.UserID, s.Title, s.StreamKey,
		s.PlaybackID)
}

// Finds a stream by its secret key
func (r *Repository) GetByKey(ctx context.Context, key string) (*Stream, error) {
	var s Stream
	err := r.db.GetContext(ctx, &s, `SELECT id, user_id, title,
	 stream_key, playback_id FROM streams
	  WHERE stream_key = $1`, key)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrStreamNotFound
	}

	if err != nil {
		return nil, err
	}

	return &s, nil
}

// Listlive function returns every stream currently live
func (r *Repo) ListLive(ctx context.Context) ([]*Stream, error) {
	streams := []*Stream{} // empty slice not nil, so JSON shows [] instead of null

	err := r.db.SelectContext(ctx, &streams,
		`SELECT id, user_id, title, stream_key, playback_id, status, started_at
	FROM streams WHERE status = 'live'
	ORDER BY started_at DESC`)
	return streams, err
}

// SetStatus switches a stream live or offline and stamps started_at when going live
func (r *Repo) SetStatus(ctx context.Context, id int, status Status) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE streams SET status = $1 WHERE id = $2`, status, id)
	return err
}
