package db

import (
	"fmt"
	"sort"
	"strings"
)

// QueryError wraps a database error with the query parameters that
// caused it, enabling richer diagnostic output on failures.
type QueryError struct {
	Params map[string]any
	Err    error
}

// Error returns the underlying error message followed by a sorted
// flat list of query parameters, e.g.:
//
//	some db error [key1=val1 key2=val2]
func (e *QueryError) Error() string {
	if len(e.Params) == 0 {
		return e.Err.Error()
	}
	keys := make([]string, 0, len(e.Params))
	for k := range e.Params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%v", k, e.Params[k]))
	}
	return fmt.Sprintf(
		"%s [%s]", e.Err.Error(), strings.Join(parts, " "),
	)
}

// Unwrap returns the wrapped error, preserving errors.Is/As chains.
func (e *QueryError) Unwrap() error {
	return e.Err
}

func newQueryError(err error, params map[string]any) *QueryError {
	return &QueryError{Err: err, Params: params}
}
