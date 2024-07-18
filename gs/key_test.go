package gs

import (
	"bytes"
	"context"
	"fmt"
	"github.com/knights-analytics/afs/option"
	"github.com/knights-analytics/afs/url"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"strings"
	"testing"
)

func TestAES256Key_SetHeader(t *testing.T) {
	jwtConfig, err := NewTestJwtConfig()
	if err != nil {
		t.Skip(err)
		return
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
			URL:         fmt.Sprintf("gs://%v/", TestBucket),
			data:        []byte("this is test"),
		},

		{
			description: "securing data with base64key",
			location:    "folder/secret2.txt",
			URL:         fmt.Sprintf("gs://%v/", TestBucket),
			data:        []byte("this is test"),
			base64Key:   "eGR4ZHhkeGR4ZHhkeGR4ZHhkeGR4ZHhkeGR4ZHhkeGQ=",
		},
	}

	mgr := New(jwtConfig)

	defer func() {
		_ = mgr.Delete(ctx, fmt.Sprintf("gs://%v/", TestBucket))
	}()
	for _, useCase := range useCases {

		var key *option.AES256Key
		if useCase.key != "" {
			key, err = option.NewAES256Key([]byte(useCase.key))
			assert.Nil(t, err, useCase.description)

		} else {
			key, err = option.NewBase64AES256Key(useCase.base64Key)
			assert.Nil(t, err, useCase.description)
		}

		URL := url.Join(useCase.URL, useCase.location)
		err := mgr.Upload(ctx, URL, 0644, bytes.NewReader(useCase.data), key)
		assert.Nil(t, err, useCase.description)
		_, err = mgr.OpenURL(ctx, URL)
		assert.NotNil(t, err, useCase.description)
		reader, err := mgr.OpenURL(ctx, URL, key)
		if !assert.Nil(t, err, useCase.description) {
			continue
		}

		data, err := ioutil.ReadAll(reader)
		assert.EqualValues(t, useCase.data, data, useCase.description)

	}

}
