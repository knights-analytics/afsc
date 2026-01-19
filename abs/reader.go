package abs

import (
	"context"
	"fmt"
	"io"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
	"github.com/viant/afs/option"
)

type reader struct {
	from     int64
	size     int
	location string
	key      *option.AES256Key
	storager *storager
	ctx      context.Context
}

func (t *reader) Seek(offset int64, whence int) (int64, error) {
	if whence != io.SeekStart {
		return 0, fmt.Errorf("unsupported whence: %v", whence)
	}
	if int(offset) > t.size {
		return 0, io.EOF
	}
	t.from = offset
	return 0, nil
}

func (t *reader) Read(dest []byte) (int, error) {
	if int(t.from) >= t.size {
		return 0, io.EOF
	}
	from := t.from
	to := int64(int(t.from) + len(dest) - 1)
	if to >= int64(t.size) {
		to = int64(t.size - 1)
	}

	o := &azblob.DownloadStreamOptions{
		Range: blob.HTTPRange{
			Offset: from,
			Count:  to - from + 1,
		},
	}

	if t.key != nil && len(t.key.Key) > 0 {
		keyStr := string(t.key.Key)
		o.CPKInfo = &blob.CPKInfo{
			EncryptionKey:       &keyStr,
			EncryptionKeySHA256: &t.key.Base64KeyMd5Hash,
			EncryptionAlgorithm: &[]blob.EncryptionAlgorithmType{blob.EncryptionAlgorithmTypeAES256}[0],
		}
	}

	resp, err := t.storager.client.DownloadStream(t.ctx, t.storager.container, t.location, o)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	readSoFar := 0
	for {
		read, err := resp.Body.Read(dest[readSoFar:])
		if read > 0 {
			readSoFar += read
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return readSoFar, err
		}
		if readSoFar >= len(dest) {
			break
		}
	}
	t.from += int64(readSoFar)
	return readSoFar, nil
}

func NewReadSeeker(ctx context.Context, storager *storager, location string, key *option.AES256Key, size int) io.ReadSeeker {
	return &reader{
		ctx:      ctx,
		storager: storager,
		location: location,
		key:      key,
		size:     size,
	}
}
