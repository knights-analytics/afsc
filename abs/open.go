package abs

import (
	"context"
	"io"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
	"github.com/viant/afs/base"
	"github.com/viant/afs/option"
	"github.com/viant/afs/storage"
)

func (s *storager) Open(ctx context.Context, location string, options ...storage.Option) (io.ReadCloser, error) {
	location = strings.Trim(location, "/")
	stream := &option.Stream{}
	key := &option.AES256Key{}
	option.Assign(options, &stream, &key)

	if stream.PartSize > 0 {
		info, err := s.Get(ctx, location, options...)
		if err != nil {
			return nil, err
		}
		stream.Size = int(info.Size())
		readSeeker := NewReadSeeker(ctx, s, location, key, int(info.Size()))
		return base.NewStreamReader(stream, readSeeker), nil
	}

	o := &azblob.DownloadStreamOptions{}
	if len(key.Key) > 0 {
		keyStr := string(key.Key)
		o.CPKInfo = &blob.CPKInfo{
			EncryptionKey:       &keyStr,
			EncryptionKeySHA256: &key.Base64KeyMd5Hash,
			EncryptionAlgorithm: &[]blob.EncryptionAlgorithmType{blob.EncryptionAlgorithmTypeAES256}[0],
		}
	}

	resp, err := s.client.DownloadStream(ctx, s.container, location, o)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}
