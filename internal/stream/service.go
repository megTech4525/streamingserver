package stream

import (
	"context"
	"errors"

	"main.go/internal/auth"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

//this part makes a stream with a secret key and an public playback id.

func (s *Service) CreateStream(ctx context.Context, userID int, title string) (*Stream, error) {
	key, err := auth.GenerateStreamKey()
	if err != nil {
		return nil, err
	}

	playbackID, err := auth.GeneratePlaybackID()
	if err != nil {
		return nil, err
	}

	st := &Stream{UserID: userID, Title: title, StreamKey: key, PlaybackID: playbackID}

	if err := s.repo.Create(ctx, st); err != nil {
		return nil, err
	}
	return st, nil
}

//ValidateKey function is called by the media server when someone starts to broadcast

func (s *Service) ValidateKey(ctx context.Context, key string) (*Stream, error) {
	st, err := s.repo.GetByKey(ctx, key)
	if errors.Is(err, ErrNotFound) {
		return nil, ErrInvalidKey
	}
	return st, err
}

func (s *Service) MarkLive(ctx context.Context, id int) error {
	return s.repo.SetStatus(ctx, id, StatusLive)
}

func (s *Service) MarkOffLine(ctx context.Context, id int) error {
	return s.repo.SetStatus(ctx, id, StatusOffline)

}

func (s *Service) ListLive(ctx context.Context) ([]*Stream, error) {
	return s.repo.ListLive(ctx)
}
