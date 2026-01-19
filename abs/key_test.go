package abs

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/viant/afs/option"
	"github.com/viant/afs/url"
)

func TestAES256Key_SetHeader(t *testing.T) {
	if os.Getenv("AZURE_STORAGE_CONNECTION_STRING") == "" && os.Getenv("AZURE_STORAGE_ACCOUNT") == "" {
		t.Skip("Azure credentials not set")
	}

	ctx := context.Background()
	var useCases = []struct {
		description string
		URL         string
		location    string
		data        []byte
		key         string
		base64Key   string
	}{
		{
			description: "securing data with key",
			key:         strings.Repeat("xd", 16),
			location:    "folder/secret1.txt",
			URL:         fmt.Sprintf("abs://%v/", TestContainer),
			data:        []byte("this is test"),
		},
		{
			description: "securing data with base64key",
			location:    "folder/secret2.txt",
			URL:         fmt.Sprintf("abs://%v/", TestContainer),
			data:        []byte("this is test"),
			base64Key:   "eGR4ZHhkeGR4ZHhkeGR4ZHhkeGR4ZHhkeGR4ZHhkeGQ=",
		},
	}

	mgr := New()
	for _, useCase := range useCases {
		var key *option.AES256Key
		var err error
		if useCase.key != "" {
			key, err = option.NewAES256Key([]byte(useCase.key))
			assert.Nil(t, err, useCase.description)
		} else {
			key, err = option.NewBase64AES256Key(useCase.base64Key)
			assert.Nil(t, err, useCase.description)
		}

		URL := url.Join(useCase.URL, useCase.location)
		err = mgr.Upload(ctx, URL, 0644, bytes.NewReader(useCase.data), key)
		assert.Nil(t, err, useCase.description)

		_, err = mgr.OpenURL(ctx, URL)
		assert.NotNil(t, err, useCase.description)

		reader, err := mgr.OpenURL(ctx, URL, key)
		if !assert.Nil(t, err, useCase.description) {
			continue
		}

		data, err := io.ReadAll(reader)
		assert.EqualValues(t, useCase.data, data, useCase.description)
	}
}
