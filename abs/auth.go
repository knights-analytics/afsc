package abs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/viant/afs/option"
	"github.com/viant/afs/storage"
)

const (
	azureStorageConnectionString = "AZURE_STORAGE_CONNECTION_STRING"
	azureStorageAccount          = "AZURE_STORAGE_ACCOUNT"
	azureStorageKey              = "AZURE_STORAGE_KEY"
)

// AuthConfig represents an azure auth config
type AuthConfig struct {
	ConnectionString string `json:",omitempty"`
	AccountName      string `json:",omitempty"`
	AccountKey       string `json:",omitempty"`
}

// ClientOptions returns client options
func (c *AuthConfig) ClientOptions() (*azblob.ClientOptions, error) {
	return nil, nil
}

// NewAuthConfig returns new auth config from location or options
func NewAuthConfig(options ...storage.Option) (*AuthConfig, error) {
	location := &option.Location{}
	var JSONPayload = make([]byte, 0)
	option.Assign(options, &location, &JSONPayload)
	if location.Path == "" && len(JSONPayload) == 0 {
		return nil, fmt.Errorf("auth location was empty")
	}
	if location.Path != "" {
		locationPath := location.Path
		if strings.HasPrefix(locationPath, "~/") {
			locationPath = path.Join(os.Getenv("HOME"), locationPath[2:])
		}
		file, err := os.Open(locationPath)
		if err != nil {
			return nil, fmt.Errorf("failed to open auth config: %w", err)
		}
		defer func() { _ = file.Close() }()
		if JSONPayload, err = io.ReadAll(file); err != nil {
			return nil, err
		}

	}
	authConfig := &AuthConfig{}
	err := json.NewDecoder(bytes.NewReader(JSONPayload)).Decode(authConfig)
	return authConfig, err
}

func filterAuthOption(options []storage.Option) (*AuthConfig, error) {
	authConfig := &AuthConfig{}
	if _, ok := option.Assign(options, &authConfig); ok {
		return authConfig, nil
	}
	return nil, nil
}

func getClient(_ context.Context, accountName string, options []storage.Option) (*azblob.Client, error) {
	connectionString := os.Getenv(azureStorageConnectionString)
	accountKey := os.Getenv(azureStorageKey)
	if accountName == "" {
		accountName = os.Getenv(azureStorageAccount)
	}

	authConfig, _ := filterAuthOption(options)
	if authConfig != nil {
		if authConfig.ConnectionString != "" {
			connectionString = authConfig.ConnectionString
		}
		if authConfig.AccountName != "" {
			accountName = authConfig.AccountName
		}
		if authConfig.AccountKey != "" {
			accountKey = authConfig.AccountKey
		}
	}

	if connectionString != "" {
		return azblob.NewClientFromConnectionString(connectionString, nil)
	}

	if accountName != "" && accountKey != "" {
		serviceURL := fmt.Sprintf("https://%s.blob.core.windows.net/", accountName)
		cred, err := azblob.NewSharedKeyCredential(accountName, accountKey)
		if err != nil {
			return nil, err
		}
		return azblob.NewClientWithSharedKeyCredential(serviceURL, cred, nil)
	}

	return nil, nil
}
