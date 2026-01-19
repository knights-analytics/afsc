package abs

import (
	"context"

	"github.com/viant/afs/storage"
)

func (s *storager) Exists(ctx context.Context, location string, options ...storage.Option) (bool, error) {
	_, err := s.Get(ctx, location, options...)
	if isNotFound(err) {
		return false, nil
	}
	return err == nil, err
}
