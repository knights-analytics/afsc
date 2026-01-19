package abs

import (
	"context"
	"strings"

	"github.com/viant/afs/storage"
)

func (s *storager) Delete(ctx context.Context, location string, options ...storage.Option) error {
	location = strings.Trim(location, "/")
	_, err := s.client.DeleteBlob(ctx, s.container, location, nil)
	if err != nil && isNotFound(err) {
		return nil
	}
	return err
}
