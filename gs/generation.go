package gs

import (
	"github.com/knights-analytics/afs/option"
	"github.com/knights-analytics/afs/storage"
)

type matcher func(generation int64)

func (s *storager) setGeneration(match, noMatch matcher, options []storage.Option) {
	generation := &option.Generation{}
	if _, ok := option.Assign(options, &generation); ok {
		if generation.WhenMatch {
			match(generation.Generation)
		} else {
			noMatch(generation.Generation)
		}
	}
}
