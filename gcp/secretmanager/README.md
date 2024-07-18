## GCP Secret Manager storage

## Usage


```go
package mypkg

import (
	"context"
	"fmt"
	"github.com/knights-analytics/afs"
	"github.com/knights-analytics/afs/file"
	_ "github.com/knights-analytics/afsc/gcp"
	"log"
	"strings"
)

func Example_DownloadWithURL() {
	fs := afs.New()
	URL := "gcp://secretmanager/projects/gcp-e2e/secrets/test2sec"
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

```
