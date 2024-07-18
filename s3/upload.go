package s3

import (
	"bytes"
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/knights-analytics/afs/option"
	"github.com/knights-analytics/afs/option/content"
	"github.com/knights-analytics/afs/storage"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

// Upload uploads content
func (s *storager) Upload(ctx context.Context, destination string, mode os.FileMode, reader io.Reader, options ...storage.Option) error {
	destination = strings.Trim(destination, "/")
	err := s.upload(ctx, destination, mode, reader, options)
	if err != nil {
		return err
	}
	return s.presign(ctx, destination, options)
}

func (s *storager) updateChecksum(input *s3.PutObjectInput, md5Hash *option.Md5, data []byte) {
	if len(md5Hash.Hash) == 0 {
		md5Hash = option.NewMd5(data)
	}
	input.ContentMD5 = aws.String(md5Hash.Encode())
}

func (s *storager) upload(ctx context.Context, destination string, mode os.FileMode, reader io.Reader, options []storage.Option) error {
	md5Hash := &option.Md5{}
	key := &option.AES256Key{}
	checksum := &option.SkipChecksum{}
	meta := &content.Meta{}
	serverSideEncryption := &option.ServerSideEncryption{}
	stream := &option.Stream{}
	grant := &option.Grant{}
	acl := &option.ACL{}
	option.Assign(options, &md5Hash, &key, &checksum, &meta, &serverSideEncryption, &stream, &grant, &acl)
	if !checksum.Skip {
		input := &s3.PutObjectInput{
			Bucket:   &s.bucket,
			Key:      aws.String(destination),
			Metadata: map[string]*string{},
		}

		updateMetaContent(meta, input)

		content, err := ioutil.ReadAll(reader)
		if err != nil {
			return err
		}
		s.updateChecksum(input, md5Hash, content)
		input.Metadata[contentMD5MetaKey] = input.ContentMD5
		input.Body = bytes.NewReader(content)

		if acl.ACL != "" {
			input.ACL = &acl.ACL
		}

		if grant.FullControl != "" {
			input.GrantFullControl = &grant.FullControl
		}
		if grant.FullControl != "" {
			input.GrantRead = &grant.Read
		}
		if grant.FullControl != "" {
			input.GrantReadACP = &grant.ReadACP
		}
		if grant.FullControl != "" {
			input.GrantWriteACP = &grant.WriteACP
		}

		if len(key.Key) > 0 {
			input.SetSSECustomerKey(string(key.Key))
			input.SetSSECustomerKeyMD5(key.Base64KeyMd5Hash)
			input.SetSSECustomerAlgorithm(customEncryptionAlgorithm)
		}

		if serverSideEncryption.Algorithm != "" {
			input.ServerSideEncryption = aws.String(serverSideEncryption.Algorithm)
		}

		_, err = s.PutObjectWithContext(ctx, input)
		if err != nil {
			if err == credentials.ErrNoValidProvidersFoundInChain {
				s.initS3Client()
			}
			if strings.Contains(err.Error(), noSuchBucketMessage) {
				if err = s.createBucket(ctx); err != nil {
					return err
				}
				input.Body = bytes.NewReader(content)
				_, err = s.PutObjectWithContext(ctx, input)
			}
		}
		if err != nil {
			err = errors.Wrapf(err, "failed to upload: s3://%v/%v", s.bucket, destination)
		}
		return err
	}
	var sess *session.Session
	if s.config == nil {
		sess = session.New()
	} else {
		sess = session.New(s.config)
	}
	uploader := s3manager.NewUploader(sess)
	if stream.PartSize > 0 {
		uploader.PartSize = int64(stream.PartSize)
	}
	input := &s3manager.UploadInput{
		Bucket:   aws.String(s.bucket),
		Key:      aws.String(destination),
		Body:     reader,
		Metadata: map[string]*string{},
	}
	if grant.FullControl != "" {
		input.GrantFullControl = &grant.FullControl
	}
	if grant.FullControl != "" {
		input.GrantRead = &grant.Read
	}
	if grant.FullControl != "" {
		input.GrantReadACP = &grant.ReadACP
	}
	if grant.FullControl != "" {
		input.GrantWriteACP = &grant.WriteACP
	}
	if acl.ACL != "" {
		input.ACL = &acl.ACL
	}

	if len(meta.Values) > 0 {
		for k := range meta.Values {
			value := meta.Values[k]
			switch k {
			case content.Type:
				input.ContentType = &value
				continue
			case content.Encoding:
				input.ContentEncoding = &value
				continue
			case content.Language:
				input.ContentLanguage = &value
				continue
			}
			input.Metadata[k] = &value
		}
	}
	_, err := uploader.Upload(input)
	if err != nil {
		return err
	}

	sizer, ok := reader.(storage.Sizer)
	if !ok {
		return nil
	}
	if objects, err := s.List(ctx, destination); err == nil && len(objects) == 1 {
		if objects[0].Size() != sizer.Size() {
			err = errors.Errorf("corrupted upload: s3://%v/%v expected size: %v, but had: %v", s.bucket, destination, sizer.Size(), objects[0].Size())
		}
	}
	return err
}

func updateMetaContent(meta *content.Meta, input *s3.PutObjectInput) {
	if len(meta.Values) > 0 {
		for k := range meta.Values {
			value := meta.Values[k]
			switch k {
			case content.Type:
				input.ContentType = &value
				continue
			case content.Encoding:
				input.ContentEncoding = &value
				continue
			case content.Language:
				input.ContentLanguage = &value
				continue
			}
			input.Metadata[k] = &value
		}
	}
}
