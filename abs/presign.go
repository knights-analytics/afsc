package abs

import (
	"context"
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/sas"
	"github.com/viant/afs/option"
	"github.com/viant/afs/storage"
)

func (s *storager) presign(ctx context.Context, destination string, options []storage.Option) error {
	preSign := &option.PreSign{}
	if _, ok := option.Assign(options, &preSign); !ok {
		return nil
	}

	expiry := time.Now().Add(preSign.TimeToLive)
	if preSign.TimeToLive == 0 {
		expiry = time.Now().Add(15 * time.Minute)
	}

	permissions := sas.BlobPermissions{
		Read: true,
	}

	if s.accountKey != "" {
		cred, err := azblob.NewSharedKeyCredential(s.accountName, s.accountKey)
		if err != nil {
			return err
		}
		signatureValues := sas.BlobSignatureValues{
			Protocol:      sas.ProtocolHTTPS,
			ExpiryTime:    expiry,
			Permissions:   permissions.String(),
			ContainerName: s.container,
			BlobName:      destination,
		}
		query, err := signatureValues.SignWithSharedKey(cred)
		if err != nil {
			return err
		}
		preSign.URL = fmt.Sprintf("https://%s.blob.core.windows.net/%s/%s?%s", s.accountName, s.container, destination, query.Encode())
		return nil
	}

	return fmt.Errorf("presign requires shared key access (AZURE_STORAGE_KEY)")
}
