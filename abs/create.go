package abs

import (
	"context"
	"io"
	"os"
	"strings"

	"github.com/viant/afs/storage"
)

func (s *storager) Create(ctx context.Context, destination string, mode os.FileMode, reader io.Reader, isDir bool, options ...storage.Option) error {
	destination = strings.Trim(destination, "/")
	if !isDir {
		return s.Upload(ctx, destination, mode, reader, options...)
	}
	return nil
}
