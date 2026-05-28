package storage

import "errors"

var (
	// ErrProviderUnavailable indicates that a provider cannot service requests.
	ErrProviderUnavailable = errors.New("storage provider unavailable")
	// ErrUnsupportedProvider indicates that provider selection requested an implementation not present in this build.
	ErrUnsupportedProvider = errors.New("unsupported storage provider")
	// ErrNotFound indicates that a requested storage record does not exist.
	ErrNotFound = errors.New("storage record not found")
)
