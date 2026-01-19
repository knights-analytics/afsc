package abs

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/viant/afs/option"
	"github.com/viant/afs/storage"
)

func (s *storager) Move(ctx context.Context, sourcePath, destContainer, destPath string, options ...storage.Option) error {
	sourcePath = strings.Trim(sourcePath, "/")
	destPath = strings.Trim(destPath, "/")
	info, err := s.Get(ctx, sourcePath, options...)

	if isNotFound(err) {
		objectOpt := &option.ObjectKind{}
		if _, ok := option.Assign(options, &objectOpt); ok && objectOpt.File {
			return err
		}
		infoList, err := s.List(ctx, sourcePath, options...)
		if err != nil {
			return err
		}
		if len(infoList) == 0 {
			return fmt.Errorf("%v: not found", sourcePath)
		}
		for i := 1; i < len(infoList); i++ {
			name := infoList[i].Name()
			if err = s.Move(ctx, path.Join(sourcePath, name), destContainer, path.Join(destPath, name), options...); err != nil {
				return err
			}
		}
		return nil
	}
	if err != nil {
		return err
	}

	err = s.Copy(ctx, sourcePath, destContainer, destPath, options...)
	if err == nil {
		if !info.IsDir() {
			err = s.Delete(ctx, sourcePath, options...)
		}
	}
	return err
}
