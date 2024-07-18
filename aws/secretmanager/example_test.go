package secretmanager_test

import (
	"context"
	"fmt"
	"github.com/knights-analytics/afs"
	"github.com/knights-analytics/afs/file"
	_ "github.com/knights-analytics/afsc/aws"
	"log"
	"strings"
)

func Example_DownloadWithURL() {
	fs := afs.New()
	URL := "aws://secretmanager/us-west-1/secret/prod/my/test2"
	err := fs.Upload(context.TODO(), URL, file.DefaultFileOsMode, strings.NewReader("test is super secret"))
	if err != nil {
		log.Fatalf("err: %v\n", err)
	}
	data, err := fs.DownloadWithURL(context.TODO(), URL)
	if err != nil {
		log.Fatalf("err: %v\n", err)
	}
	fmt.Printf("%s %v\n", data, err)
}
