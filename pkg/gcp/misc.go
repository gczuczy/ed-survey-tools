package gcp

import (
	"time"
	"google.golang.org/api/googleapi"
	"github.com/gczuczy/ed-survey-tools/pkg/log"
)

func RateLimit[T any](f func()(T, error), wait time.Duration) (T, error) {
	l := log.GetLogger("RateLimit")
	for {
		ret, err := f()

		if err != nil {
			if gerr, ok := err.(*googleapi.Error); ok && gerr.Code == 429 {
				l.Warn().Msgf("Rate limited, Sleeping 30 secs: %T", f)
				time.Sleep(wait)
				continue
			}
			return ret, err
		} else {
			return ret, err
		}
	}
}
