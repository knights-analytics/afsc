package abs

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/viant/afs/asset"
	"github.com/viant/afs/matcher"
	"github.com/viant/afs/option"
	"github.com/viant/afs/storage"
)

var TestContainer = fmt.Sprintf("afsc-test-%v", time.Now().Format("20060102150405"))

func TestStorager_List(t *testing.T) {
	if os.Getenv("AZURE_STORAGE_CONNECTION_STRING") == "" && os.Getenv("AZURE_STORAGE_ACCOUNT") == "" {
		t.Skip("Azure credentials not set")
	}

	ctx := context.Background()
	var useCases = []struct {
		description string
		URL         string
		listURL     string
		assets      []*asset.Resource
		options     []storage.Option
		expect      []string
	}{
		{
			description: "single asset list",
			URL:         fmt.Sprintf("abs://%v/", TestContainer),
			assets: []*asset.Resource{
				asset.NewFile("list01/asset1.txt", []byte("test is test 1 "), 0655),
			},
			listURL: fmt.Sprintf("abs://%v/list01", TestContainer),
			expect:  []string{"list01", "asset1.txt"},
		},
		{
			description: "multi asset list",
			URL:         fmt.Sprintf("abs://%v/", TestContainer),
			assets: []*asset.Resource{
				asset.NewFile("list02/asset3.txt", []byte("test is test 1 "), 0655),
				asset.NewFile("list02/folder1/asset1.txt", []byte("test is test 2"), 0655),
				asset.NewFile("list02/folder1/asset2.txt", []byte("test is test 3"), 0655),
			},
			listURL: fmt.Sprintf("abs://%v/list02", TestContainer),
			expect:  []string{"list02", "asset3.txt", "folder1/asset1.txt", "folder1/asset2.txt"}, // Flat listing will return full paths relative to container?
		},
		{
			description: "multi asset list with matcher",
			URL:         fmt.Sprintf("abs://%v/", TestContainer),
			assets: []*asset.Resource{
				asset.NewFile("list03/asset1.txt", []byte("test is test 1 "), 0655),
				asset.NewFile("list03/asset2.json", []byte("test is test 1 "), 0655),
				asset.NewFile("list03/asset3.csv", []byte("test is test 1 "), 0655),
			},
			options: []storage.Option{
				&matcher.Basic{Suffix: ".json"},
			},
			listURL: fmt.Sprintf("abs://%v/list03", TestContainer),
			expect:  []string{"asset2.json"},
		},
	}

	mgr := New()
	// Container must exist. In real tests we might need to create it.
	// defer mgr.Delete(ctx, fmt.Sprintf("abs://%v/", TestContainer))

	for _, useCase := range useCases {
		err := asset.Create(mgr, useCase.URL, useCase.assets)
		if !assert.Nil(t, err, useCase.description) {
			continue
		}
		objects, err := mgr.List(ctx, useCase.listURL, useCase.options...)
		if !assert.Nil(t, err, useCase.description) {
			continue
		}

		assert.Equal(t, len(useCase.expect), len(objects), useCase.description)
		actuals := make(map[string]bool)
		for _, object := range objects {
			actuals[object.Name()] = true
		}
		for _, expect := range useCase.expect {
			ok := actuals[expect]
			assert.True(t, ok, useCase.description+" "+expect)
		}
	}
}

func TestStorager_Upload(t *testing.T) {
	if os.Getenv("AZURE_STORAGE_CONNECTION_STRING") == "" && os.Getenv("AZURE_STORAGE_ACCOUNT") == "" {
		t.Skip("Azure credentials not set")
	}

	ctx := context.Background()
	mgr := New()
	destURL := fmt.Sprintf("abs://%v/upload/test.txt", TestContainer)
	payload := []byte("hello azure")
	err := asset.Create(mgr, destURL, []*asset.Resource{
		asset.NewFile("test.txt", payload, 0644),
	})
	assert.Nil(t, err)

	objects, err := mgr.List(ctx, destURL)
	assert.Nil(t, err)
	assert.True(t, len(objects) > 0)
}

func TestStorager_Move(t *testing.T) {
	if os.Getenv("AZURE_STORAGE_CONNECTION_STRING") == "" && os.Getenv("AZURE_STORAGE_ACCOUNT") == "" {
		t.Skip("Azure credentials not set")
	}

	ctx := context.Background()
	mgr := New()
	sourceURL := fmt.Sprintf("abs://%v/move/source.txt", TestContainer)
	destURL := fmt.Sprintf("abs://%v/move/dest.txt", TestContainer)
	payload := []byte("move me")

	err := asset.Create(mgr, sourceURL, []*asset.Resource{
		asset.NewFile("source.txt", payload, 0644),
	})
	assert.Nil(t, err)

	err = mgr.(storage.Mover).Move(ctx, sourceURL, destURL)
	assert.Nil(t, err)

	// Check dest exists
	objects, err := mgr.List(ctx, destURL)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(objects))

	// Check source is gone
	objects, err = mgr.List(ctx, sourceURL)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(objects))
}

func TestStorager_Auth(t *testing.T) {
	if os.Getenv("AZURE_STORAGE_CONNECTION_STRING") == "" && os.Getenv("AZURE_STORAGE_ACCOUNT") == "" {
		t.Skip("Azure credentials not set")
	}

	accountName := os.Getenv("AZURE_STORAGE_ACCOUNT")
	accountKey := os.Getenv("AZURE_STORAGE_KEY")
	if accountKey == "" {
		t.Skip("AZURE_STORAGE_KEY not set")
	}

	authConfig := &AuthConfig{
		AccountName: accountName,
		AccountKey:  accountKey,
	}

	mgr := New(authConfig)
	destURL := fmt.Sprintf("abs://%v/auth/test.txt", TestContainer)
	err := asset.Create(mgr, destURL, []*asset.Resource{
		asset.NewFile("test.txt", []byte("auth test"), 0644),
	})
	assert.Nil(t, err)

	// Test IsAuthChanged
	st, err := mgr.(interface {
		Storager(ctx context.Context, URL string, options ...storage.Option) (storage.Storager, error)
	}).Storager(context.Background(), destURL)
	assert.Nil(t, err)
	absStorager := st.(*storager)
	assert.False(t, absStorager.IsAuthChanged(nil))
	assert.False(t, absStorager.IsAuthChanged([]storage.Option{authConfig}))

	newAuthConfig := &AuthConfig{AccountName: "other"}
	assert.True(t, absStorager.IsAuthChanged([]storage.Option{newAuthConfig}))
}

func TestStorager_PreSign(t *testing.T) {
	if os.Getenv("AZURE_STORAGE_ACCOUNT") == "" || os.Getenv("AZURE_STORAGE_KEY") == "" {
		t.Skip("Azure credentials not set")
	}

	ctx := context.Background()
	mgr := New()
	destURL := fmt.Sprintf("abs://%v/presign/test.txt", TestContainer)
	payload := []byte("presign me")
	err := asset.Create(mgr, destURL, []*asset.Resource{
		asset.NewFile("test.txt", payload, 0644),
	})
	assert.Nil(t, err)

	preSign := &option.PreSign{TimeToLive: time.Hour}
	err = mgr.(interface {
		PreSign(ctx context.Context, URL string, options ...storage.Option) error
	}).PreSign(ctx, destURL, preSign)
	assert.Nil(t, err)
	assert.NotEmpty(t, preSign.URL)
	fmt.Printf("PreSigned URL: %v\n", preSign.URL)
}
