package auth

import "errors"

var ErrKeyNotFound  = errors.New("api key not found or revoked")
var ErrInvalidKey   = errors.New("invalid api key format")
var ErrUnauthorized = errors.New("unauthorized")
