package gcp

import "github.com/knights-analytics/afs/storage"

// Provider returns a google storage manager
func Provider(options ...storage.Option) (storage.Manager, error) {
	return New(options...), nil
}
