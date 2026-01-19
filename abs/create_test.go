package abs

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/viant/afs/asset"
	"github.com/viant/afs/url"
)

func TestStorager_Create(t *testing.T) {
	if os.Getenv("AZURE_STORAGE_CONNECTION_STRING") == "" && os.Getenv("AZURE_STORAGE_ACCOUNT") == "" {
		t.Skip("Azure credentials not set")
	}

	ctx := context.Background()
	var useCases = []struct {
		description string
		URL         string
		assets      []*asset.Resource
	}{
		{
			description: "single asset create",
			URL:         fmt.Sprintf("abs://%v/001_create/", TestContainer),
			assets: []*asset.Resource{
				asset.NewFile("asset1.txt", []byte("test is test"), 0655),
			},
		},
		{
			description: "multi asset create",
			URL:         fmt.Sprintf("abs://%v/002_create/", TestContainer),
			assets: []*asset.Resource{
				asset.NewFile("folder1/asset1.txt", []byte("test is test"), 0655),
				asset.NewFile("folder1/asset2.txt", []byte("test is test"), 0655),
			},
		},
	}

	mgr := New()
	for _, useCase := range useCases {
		for _, uasset := range useCase.assets {
			var reader io.Reader
			if len(uasset.Data) > 0 {
				reader = bytes.NewReader(uasset.Data)
			}

			err := mgr.Create(ctx, url.Join(useCase.URL, uasset.Name), 0644, uasset.Dir, reader)
			assert.Nil(t, err, useCase.description)
		}
		actuals, err := asset.Load(mgr, useCase.URL)
		if !assert.Nil(t, err, useCase.description) {
			continue
		}
		for _, uasset := range useCase.assets {
			actual, ok := actuals[uasset.Name]
			assert.True(t, ok, useCase.description+" "+uasset.Name)
			assert.NotNil(t, actual, useCase.description+" "+uasset.Name)
		}
	}
}
