package auth

import "errors"

var ErrAuthenticationRequired = errors.New("authentication required")
var ErrAuthenticationFailed = errors.New("authentication failed")
