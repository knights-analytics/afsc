package abs

import (
	"context"
	"os"
	"path"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/viant/afs/file"
	"github.com/viant/afs/option"
	"github.com/viant/afs/storage"
)

func (s *storager) List(ctx context.Context, location string, options ...storage.Option) ([]os.FileInfo, error) {
	location = strings.Trim(location, "/")
	matcher, _ := option.GetListOptions(options)
	var result = make([]os.FileInfo, 0)

	// Add the directory itself
	info, err := s.Get(ctx, location, options...)
	if err == nil {
		if matcher("", info) {
			result = append(result, info)
		}
	} else if !isNotFound(err) {
		return nil, err
	}

	prefix := location
	if prefix != "" && !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}

	pager := s.client.NewListBlobsFlatPager(s.container, &azblob.ListBlobsFlatOptions{
		Prefix: &prefix,
	})

	for pager.More() {
		resp, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, blobItem := range resp.Segment.BlobItems {
			if blobItem.Name == nil {
				continue
			}
			if *blobItem.Name == prefix || *blobItem.Name == location {
				continue
			}

			// Azure doesn't have a built-in concept of directories in flat listing.
			// We might need to handle virtual directories if we want Hierarchical listing.
			// But NewListBlobsHierarchyPager seems to be what was intended.
			// If it's missing, we use Flat and manual parsing or check if NewListBlobsHierarchyPager exists in another package.

			_, name := path.Split(*blobItem.Name)
			size := int64(0)
			if blobItem.Properties != nil && blobItem.Properties.ContentLength != nil {
				size = *blobItem.Properties.ContentLength
			}
			modTime := time.Now()
			if blobItem.Properties != nil && blobItem.Properties.LastModified != nil {
				modTime = *blobItem.Properties.LastModified
			}

			info := file.NewInfo(name, size, file.DefaultFileOsMode, modTime, false, blobItem)
			if matcher(location, info) {
				result = append(result, info)
			}
		}
	}

	return result, nil
}
