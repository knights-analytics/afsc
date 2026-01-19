package abs

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/viant/afs"
	"github.com/viant/afs/asset"
	"github.com/viant/afs/url"
)

func TestStorager_Copy(t *testing.T) {
	if os.Getenv("AZURE_STORAGE_CONNECTION_STRING") == "" && os.Getenv("AZURE_STORAGE_ACCOUNT") == "" {
		t.Skip("Azure credentials not set")
	}

	ctx := context.Background()
	var useCases = []struct {
		description string
		URL         string
		source      string
		dest        string
		assets      []*asset.Resource
	}{
		{
			description: "single asset copy",
			URL:         fmt.Sprintf("abs://%v/", TestContainer),
			dest:        "copy001/dst",
			source:      "copy001/src",
			assets: []*asset.Resource{
				asset.NewFile("copy001/src/asset1.txt", []byte("test is test 1 "), 0655),
			},
		},
		{
			description: "multi asset copy",
			URL:         fmt.Sprintf("abs://%v/", TestContainer),
			dest:        "copy002/dst",
			source:      "copy002/src",
			assets: []*asset.Resource{
				asset.NewFile("copy002/src/folder1/asset1.txt", []byte("test is test 2"), 0655),
				asset.NewFile("copy002/src/folder1/asset2.txt", []byte("test is test 3"), 0655),
			},
		},
	}
	fs := afs.New()
	mgr := New()

	for _, useCase := range useCases {
		err := asset.Create(mgr, useCase.URL, useCase.assets)
		assert.Nil(t, err, useCase.description)
		err = fs.Copy(ctx, url.Join(useCase.URL, useCase.source), url.Join(useCase.URL, useCase.dest))
		assert.Nil(t, err, useCase.description)
		for _, uasset := range useCase.assets {
			URL := url.Join(useCase.URL, uasset.Name)
			URL = strings.Replace(URL, useCase.source, useCase.dest, 1)
			reader, err := mgr.OpenURL(ctx, URL)
			if !assert.Nil(t, err, useCase.description) {
				continue
			}
			data, err := io.ReadAll(reader)
			assert.EqualValues(t, uasset.Data, data, useCase.description+" "+uasset.Name)
		}
	}
}
