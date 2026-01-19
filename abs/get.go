package abs

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/viant/afs/file"
	"github.com/viant/afs/storage"
)

func (s *storager) Get(ctx context.Context, location string, options ...storage.Option) (os.FileInfo, error) {
	location = strings.Trim(location, "/")
	if location == "" {
		return file.NewInfo("/", 0, file.DefaultDirOsMode, time.Now(), true, nil), nil
	}
	resp, err := s.client.ServiceClient().NewContainerClient(s.container).NewBlobClient(location).GetProperties(ctx, nil)
	if err != nil {
		if isNotFound(err) {
			// Check if it's a "directory" by listing with prefix
			pager := s.client.NewListBlobsFlatPager(s.container, &azblob.ListBlobsFlatOptions{
				Prefix: &location,
			})
			if pager.More() {
				page, pageErr := pager.NextPage(ctx)
				if pageErr == nil && len(page.Segment.BlobItems) > 0 {
					_, name := path.Split(location)
					return file.NewInfo(name, 0, file.DefaultDirOsMode, time.Now(), true, nil), nil
				}
			}
			return nil, fmt.Errorf("%v: %w", location, os.ErrNotExist)
		}
		return nil, err
	}

	size := int64(0)
	if resp.ContentLength != nil {
		size = *resp.ContentLength
	}
	modTime := time.Now()
	if resp.LastModified != nil {
		modTime = *resp.LastModified
	}
	_, name := path.Split(location)
	return file.NewInfo(name, size, file.DefaultFileOsMode, modTime, false, resp), nil
}
