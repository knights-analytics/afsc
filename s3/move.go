package s3

import (
	"context"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/pkg/errors"
	"github.com/viant/afs/option"
	"github.com/viant/afs/storage"
)

func (s *Storager) Move(ctx context.Context, sourcePath, destBucket, destPath string, options ...storage.Option) error {
	started := time.Now()
	defer func() {
		s.logF("s3:Move %v-> %v/%v %s\n", sourcePath, destBucket, destPath, time.Since(started))
	}()
	sourcePath = strings.Trim(sourcePath, "/")
	destPath = strings.Trim(destPath, "/")
	source, err := s.get(ctx, sourcePath, options)
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
		copyer := newCopyer(s.Client, source, defaultPartSize, copyInput)
		err = copyer.copy(ctx)
	} else {
		_, err = s.Client.CopyObject(ctx, copyInput)
	}
	if err != nil {
		err = errors.Wrapf(err, "failed to move: s3://%v/%v to s3://%v/%v", s.bucket, sourcePath, destBucket, destPath)
	}
	if err == nil {
		err = s.Delete(ctx, sourcePath)
	}
	return err
}
