package abs

import (
	"context"
	"fmt"
	"strings"

	"github.com/viant/afs/storage"
)

func (s *storager) Copy(ctx context.Context, sourcePath, destContainer, destPath string, options ...storage.Option) error {
	sourcePath = strings.Trim(sourcePath, "/")
	destPath = strings.Trim(destPath, "/")

	sourceURL := fmt.Sprintf("https://%s.blob.core.windows.net/%s/%s", s.accountName, s.container, sourcePath)

	containerClient := s.client.ServiceClient().NewContainerClient(destContainer)
	blobClient := containerClient.NewBlobClient(destPath)
	_, err := blobClient.StartCopyFromURL(ctx, sourceURL, nil)
	return err
}
