package common

import "errors"

var (
	ConcurrencyUpdateError = errors.New("Concurrency update error: record timestamp has changed")
)
