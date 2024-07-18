package s3

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/knights-analytics/afs/option"
	"github.com/knights-analytics/afs/storage"
	"github.com/pkg/errors"
	"path"
	"strings"
)

const (
	maxCopySize = 5 * 1024 * 1024 * 1024
)

func (s *storager) Copy(ctx context.Context, sourcePath, destBucket, destPath string, options ...storage.Option) error {
	sourcePath = strings.Trim(sourcePath, "/")
	destPath = strings.Trim(destPath, "/")
	_, err := s.get(ctx, sourcePath, options)
	source, _ := s.get(ctx, sourcePath, nil)
	if isNotFound(err) {
		objectOpt := &option.ObjectKind{}
		if _, ok := option.Assign(options, &objectOpt); ok && objectOpt.File {
			return err
		}
		infoList, err := s.List(ctx, sourcePath)
		if err != nil {
			return err
		}
		if len(infoList) == 0 {
			return fmt.Errorf("%v: not found", sourcePath)
		}
		for i := 1; i < len(infoList); i++ {
			name := infoList[i].Name()
			if err = s.Move(ctx, path.Join(sourcePath, name), destBucket, path.Join(destPath, name)); err != nil {
				return err
			}
		}
		return nil
	}

	if err != nil {
		return err
	}

	copyInput := &s3.CopyObjectInput{
		CopySource: aws.String(s.bucket + "/" + sourcePath),
		Key:        &destPath,
		Bucket:     &destBucket,
	}
	if source.Size() >= maxCopySize {
		copyer := newCopyer(s.S3, source, defaultPartSize, copyInput)
		return copyer.copy(ctx)
	}

	_, err = s.S3.CopyObjectWithContext(ctx, copyInput)
	if err != nil {
		err = errors.Wrapf(err, "failed to copy: s3://%v/%v to s3://%v/%v", s.bucket, sourcePath, destBucket, destPath)
	}
	return err
}
