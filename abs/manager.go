package abs

import (
	"context"
	"fmt"
	"github.com/viant/afs/base"
	"github.com/viant/afs/option"
	"github.com/viant/afs/storage"
	"github.com/viant/afs/url"
)

type Manager interface {
	storage.Manager
}

type manager struct {
	*base.Manager
}

func (m *manager) provider(ctx context.Context, baseURL string, options ...storage.Option) (storage.Storager, error) {
	options = m.Options(options)
	return newStorager(ctx, baseURL, options...)
}

func (m *manager) copyInMemory(ctx context.Context, sourceURL, destURL string, options []storage.Option) error {
	objects, err := m.List(ctx, sourceURL, options...)
	if err != nil {
		return fmt.Errorf("copy source not found %v: %w", sourceURL, err)
	}
	if len(objects) == 0 {
		return fmt.Errorf("copy source not found %v", sourceURL)
	}

	downloadOptions := append(options, option.NewStream(32*1024*1024, int(objects[0].Size())))
	reader, err := m.OpenURL(ctx, sourceURL, downloadOptions...)
	if err != nil {
		return fmt.Errorf("failed open %v for copy %v: %w", sourceURL, destURL, err)
	}
	defer reader.Close()
	return m.Upload(ctx, destURL, objects[0].Mode(), reader, options...)
}

func (m *manager) Copy(ctx context.Context, sourceURL, destURL string, options ...storage.Option) error {
	absStorager, err := m.Storager(ctx, sourceURL, options)
	if err != nil {
		return err
	}
	rawStorager, ok := absStorager.(*storager)
	if !ok {
		return fmt.Errorf("expected: %T, but had: %T", rawStorager, absStorager)
	}

	sourcePath := url.Path(sourceURL)
	destContainer := url.Host(destURL)
	destPath := url.Path(destURL)

	key := &option.AES256Key{}
	_, hasKey := option.Assign(options, &key)

	if !hasKey {
		err = rawStorager.Copy(ctx, sourcePath, destContainer, destPath, options...)
	}

	if hasKey || (err != nil && sourcePath != "" && destContainer != "") {
		return m.copyInMemory(ctx, sourceURL, destURL, options)
	}

	return err
}

func (m *manager) Move(ctx context.Context, sourceURL, destURL string, options ...storage.Option) error {
	absStorager, err := m.Storager(ctx, sourceURL, options)
	if err != nil {
		return err
	}
	rawStorager, ok := absStorager.(*storager)
	if !ok {
		return fmt.Errorf("expected: %T, but had: %T", rawStorager, absStorager)
	}

	sourcePath := url.Path(sourceURL)
	destContainer := url.Host(destURL)
	destPath := url.Path(destURL)

	key := &option.AES256Key{}
	_, hasKey := option.Assign(options, &key)

	if !hasKey {
		err = rawStorager.Move(ctx, sourcePath, destContainer, destPath, options...)
	}

	if hasKey || (err != nil && sourcePath != "" && destContainer != "") {
		err = m.copyInMemory(ctx, sourceURL, destURL, options)
		if err == nil {
			err = m.Delete(ctx, sourceURL)
		}
	}

	return err
}

func newManager(options ...storage.Option) *manager {
	result := &manager{}
	baseMgr := base.New(result, Scheme, result.provider, options)
	result.Manager = baseMgr
	return result
}

// New creates abs manager
func New(options ...storage.Option) storage.Manager {
	return newManager(options...)
}
