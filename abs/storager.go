package abs

import (
	"context"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/viant/afs/storage"
	"github.com/viant/afs/url"
	"os"
)

type storager struct {
	client      *azblob.Client
	container   string
	accountName string
	accountKey  string
	config      *AuthConfig
}

func (s *storager) Close() error {
	return nil
}

func (s *storager) Scheme() string {
	return Scheme
}

func (s *storager) Config() interface{} {
	return s.config
}

func (s *storager) FilterAuthOptions(options []storage.Option) []storage.Option {
	var authOptions = make([]storage.Option, 0)
	if config, _ := filterAuthOption(options); config != nil {
		authOptions = append(authOptions, config)
	}
	return authOptions
}

// IsAuthChanged return true if auth has changes
func (s *storager) IsAuthChanged(options []storage.Option) bool {
	authOptions := s.FilterAuthOptions(options)
	if len(authOptions) == 0 {
		return false
	}
	config, _ := filterAuthOption(authOptions)
	if config == nil || s.config == nil {
		return true
	}

	return config.AccountName != s.config.AccountName ||
		config.AccountKey != s.config.AccountKey ||
		config.ConnectionString != s.config.ConnectionString
}

func newStorager(ctx context.Context, baseURL string, options ...storage.Option) (*storager, error) {
	container := url.Host(baseURL)
	if container == "" {
		return nil, fmt.Errorf("container was empty, URL: %v", baseURL)
	}

	authConfig, _ := filterAuthOption(options)
	accountName := os.Getenv("AZURE_STORAGE_ACCOUNT")
	accountKey := os.Getenv("AZURE_STORAGE_KEY")

	if authConfig != nil {
		if authConfig.AccountName != "" {
			accountName = authConfig.AccountName
		}
		if authConfig.AccountKey != "" {
			accountKey = authConfig.AccountKey
		}
	}

	client, err := getClient(ctx, accountName, options)
	if err != nil {
		return nil, err
	}

	if client == nil {
		connectionString := os.Getenv("AZURE_STORAGE_CONNECTION_STRING")
		if authConfig != nil && authConfig.ConnectionString != "" {
			connectionString = authConfig.ConnectionString
		}

		azOptions := &azblob.ClientOptions{}
		if connectionString != "" {
			client, err = azblob.NewClientFromConnectionString(connectionString, azOptions)
		} else {
			// Fallback to default Azure credential
			cred, credErr := azidentity.NewDefaultAzureCredential(nil)
			if credErr != nil {
				return nil, credErr
			}
			if accountName == "" {
				return nil, fmt.Errorf("AZURE_STORAGE_ACCOUNT environment variable is not set")
			}
			serviceURL := fmt.Sprintf("https://%s.blob.core.windows.net/", accountName)
			client, err = azblob.NewClient(serviceURL, cred, azOptions)
		}
	}

	if err != nil {
		return nil, err
	}

	return &storager{
		client:      client,
		container:   container,
		accountName: accountName,
		accountKey:  accountKey,
		config:      authConfig,
	}, nil
}

func NewStorager(ctx context.Context, baseURL string, options ...storage.Option) (storage.Storager, error) {
	return newStorager(ctx, baseURL, options...)
}
