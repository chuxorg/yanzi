package storage

import "errors"

var (
	// ErrProviderUnavailable indicates that a provider cannot service requests.
	ErrProviderUnavailable = errors.New("storage provider unavailable")
	// ErrUnsupportedProvider indicates that provider selection requested an implementation not present in this build.
	ErrUnsupportedProvider = errors.New("unsupported storage provider")
)
