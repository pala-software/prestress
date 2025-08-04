package auth

type authenticationError struct {
	message string
}

func (err authenticationError) Error() string {
	return err.message
}

func (err authenticationError) Message() string {
	return err.message
}

func (authenticationError) Status() int {
	return 401
}

var ErrAuthenticationRequired = &authenticationError{message: "authentication required"}
var ErrAuthenticationFailed = &authenticationError{message: "authentication failed"}
