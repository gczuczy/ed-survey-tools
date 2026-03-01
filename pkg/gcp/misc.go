package gcp

import (
	"fmt"
	"time"
	"google.golang.org/api/googleapi"
)

func RateLimit[T any](f func()(T, error), wait time.Duration) (T, error) {
	for {
		ret, err := f()

		if err != nil {
			if gerr, ok := err.(*googleapi.Error); ok && gerr.Code == 429 {
				fmt.Printf("Rate limited, Sleeping 30 secs\n")
				time.Sleep(wait)
				continue
			}
			return ret, err
		} else {
			return ret, err
		}
	}
}
