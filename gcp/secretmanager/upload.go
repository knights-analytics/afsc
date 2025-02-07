package secretmanager

import (
	"context"
	"io"
	"os"

	"github.com/viant/afs/storage"
	"google.golang.org/genproto/googleapis/cloud/secretmanager/v1"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"
)

// Upload uploads
func (s *Storager) Upload(ctx context.Context, destination string, mode os.FileMode, reader io.Reader, options ...storage.Option) error {
	hasSecret, _ := s.Exists(ctx, destination)
	resource, err := newResource(destination)
	if err != nil {
		return err
	}
	var secret *secretmanager.Secret
	if !hasSecret {
		secret, err = s.createSecret(ctx, resource)
	} else {
		secret, err = s.client.GetSecret(ctx, &secretmanagerpb.GetSecretRequest{Name: resource.Name()})
	}
	if err != nil {
		return err
	}
	data, err := io.ReadAll(reader)
	if err != nil {
		return err
	}
	addSecretVersionReq := &secretmanagerpb.AddSecretVersionRequest{
		Parent: secret.Name,
		Payload: &secretmanagerpb.SecretPayload{
			Data: data,
		},
	}
	_, err = s.client.AddSecretVersion(ctx, addSecretVersionReq)
	return err
}
