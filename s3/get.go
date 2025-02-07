package s3

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/pkg/errors"
	"github.com/viant/afs/file"
	"github.com/viant/afs/option"
	"github.com/viant/afs/option/content"
	"github.com/viant/afs/storage"
)

// Get returns an object for supplied location
func (s *Storager) get(ctx context.Context, location string, options []storage.Option) (os.FileInfo, error) {
	location = strings.Trim(location, "/")
	_, name := path.Split(location)

	object, err := s.HeadObject(context.Background(), &s3.HeadObjectInput{Bucket: &s.bucket,
		Key: &location})

	// object, err := s.GetObject(&s3.GetObjectInput{
	//	Bucket: &s.bucket,
	//	Key:    &location,
	// })
	if err != nil {
		return nil, err
	}
	hasObject := object != nil && (object.ContentLength != nil || object.LastModified != nil)
	if !hasObject {
		return nil, fmt.Errorf(noSuchKeyMessage + " " + location)
	}
	s.assignMetadataWithHead(options, object)
	contentLength := int64(0)
	modified := time.Now()
	if object.LastModified != nil {
		modified = *object.LastModified
	}
	if object.ContentLength != nil {
		contentLength = *object.ContentLength
	}
	if err = s.presign(ctx, location, options); err != nil {
		return nil, err
	}
	return file.NewInfo(name, contentLength, file.DefaultFileOsMode, modified, false, object), nil
}

func (s *Storager) assignMetadata(options []storage.Option, object *s3.GetObjectOutput) {
	meta := &content.Meta{}
	if _, ok := option.Assign(options, &meta); ok {
		meta.Values = make(map[string]string)
		if len(object.Metadata) > 0 {
			for k, v := range object.Metadata {
				value := ""
				if v != "" {
					value = v
				}
				meta.Values[k] = value
			}
		}
	}
}

func (s *Storager) assignMetadataWithHead(options []storage.Option, object *s3.HeadObjectOutput) {
	meta := &content.Meta{}
	if _, ok := option.Assign(options, &meta); ok {
		meta.Values = make(map[string]string)
		if len(object.Metadata) > 0 {
			for k, v := range object.Metadata {
				value := ""
				if v != "" {
					value = v
				}
				meta.Values[k] = value
			}
		}
	}
}

// Get returns an object for supplied location
func (s *Storager) Get(ctx context.Context, location string, options ...storage.Option) (os.FileInfo, error) {
	started := time.Now()
	defer func() {
		s.logF("s3:Get %v %s\n", location, time.Since(started))
	}()
	location = strings.Trim(location, "/")
	info, err := s.get(ctx, location, options)
	if err == nil {
		return info, err
	}
	if isNotFound(err) {
		objectKind := &option.ObjectKind{}
		if _, ok := option.Assign(options, &objectKind); ok && objectKind.File {
			return nil, err
		}
	}
	options = append(options, option.NewPage(0, 1))
	objects, err := s.List(ctx, location, options...)
	if err != nil {
		return nil, err
	}
	if len(objects) == 0 {
		return nil, errors.Errorf("%v %v", location, doesNotExistsMessage)
	}
	return objects[0], err
}
