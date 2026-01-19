package abs

import (
	"context"
	"io"
	"os"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
	"github.com/viant/afs/option"
	"github.com/viant/afs/storage"
)

func (s *storager) Upload(ctx context.Context, destination string, mode os.FileMode, reader io.Reader, options ...storage.Option) error {
	destination = strings.Trim(destination, "/")
	key := &option.AES256Key{}
	option.Assign(options, &key)

	o := &azblob.UploadStreamOptions{}
	if len(key.Key) > 0 {
		keyStr := string(key.Key)
		o.CPKInfo = &blob.CPKInfo{
			EncryptionKey:       &keyStr,
			EncryptionKeySHA256: &key.Base64KeyMd5Hash,
			EncryptionAlgorithm: &[]blob.EncryptionAlgorithmType{blob.EncryptionAlgorithmTypeAES256}[0],
		}
	}

	_, err := s.client.UploadStream(ctx, s.container, destination, reader, o)
	return err
}
