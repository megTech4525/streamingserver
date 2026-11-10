package stream

import "time"

type Status string

const (
	StatusOffline Status = "offline"
	StatusLive    Status = "live"
)

type Stream struct {
	ID         int        `db:"id" json:"id"`
	UserID     int        `db:"user_id" json:"user_id"`
	Title      string     `db:"title" json:"title"`
	StreamKey  string     `db:"stream_key" json:"-"`
	PlaybackID string     `db:"playback_id" json:"playback_id"`
	Status     Status     `db:"status" json:"status"`
	StartedAt  *time.Time `db:"started_at" json:"started_at"`
}
