package secretmanager

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/viant/afs/storage"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"
)

// Open returns a reader closer for supplied resources
func (s *Storager) Open(ctx context.Context, resourceID string, options ...storage.Option) (io.ReadCloser, error) {
	resource, err := newResource(resourceID)
	if err != nil {
		return nil, err
	}
	if resource.Secret == "" {
		return nil, fmt.Errorf("secret was empty")
	}
	accessRequest := &secretmanagerpb.AccessSecretVersionRequest{
		Name: resource.VersionedName(),
	}
	result, err := s.client.AccessSecretVersion(ctx, accessRequest)
	if err != nil {
		return nil, err
	}
	reader := bytes.NewReader(result.Payload.Data)
	return io.NopCloser(reader), nil
}
