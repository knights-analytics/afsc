package gcp

import (
	"github.com/knights-analytics/afs"
)

func init() {
	afs.GetRegistry().Register(Scheme, Provider)
}
