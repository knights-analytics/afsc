package s3

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/knights-analytics/afs/file"
	"github.com/knights-analytics/afs/option"
	"github.com/knights-analytics/afs/option/content"
	"github.com/knights-analytics/afs/storage"
	"github.com/pkg/errors"
	"os"
	"path"
	"strings"
	"time"
)

// Get returns an object for supplied location
func (s *storager) get(ctx context.Context, location string, options []storage.Option) (os.FileInfo, error) {
	location = strings.Trim(location, "/")
	_, name := path.Split(location)

	object, err := s.HeadObject(&s3.HeadObjectInput{Bucket: &s.bucket,
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

func (s *storager) assignMetadata(options []storage.Option, object *s3.GetObjectOutput) {
	meta := &content.Meta{}
	if _, ok := option.Assign(options, &meta); ok {
		meta.Values = make(map[string]string)
		if len(object.Metadata) > 0 {
			for k, v := range object.Metadata {
				value := ""
				if v != nil {
					value = *v
				}
				meta.Values[k] = value
			}
		}
	}
}

func (s *storager) assignMetadataWithHead(options []storage.Option, object *s3.HeadObjectOutput) {
	meta := &content.Meta{}
	if _, ok := option.Assign(options, &meta); ok {
		meta.Values = make(map[string]string)
		if len(object.Metadata) > 0 {
			for k, v := range object.Metadata {
				value := ""
				if v != nil {
					value = *v
				}
				meta.Values[k] = value
			}
		}
	}
}

// Get returns an object for supplied location
func (s *storager) Get(ctx context.Context, location string, options ...storage.Option) (os.FileInfo, error) {
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
