package prestress

import (
	"fmt"
	"net/http"
)

// TODO: Make default roles configurable
const AnonymousRole = "anonymous"
const AuthenticatedRole = "authenticated"

type AuthenticationResult struct {
	// Maybe be empty if user is not recognized (anonymous).
	Variables map[string]any

	// Always set to some role name. It's anonymous if user is not authenticated.
	Role string
}

// Details for logging
func (result AuthenticationResult) Details() map[string]string {
	details := make(map[string]string, len(result.Variables)+1)
	details["role"] = result.Role
	for key, val := range result.Variables {
		details[key] = fmt.Sprint(val)
	}
	return details
}

type Authenticator interface {
	// If user can be authenticated from the request, pointer to result is returned.
	// Otherwise response is written and nil is returned.
	Authenticate(http.ResponseWriter, *http.Request) *AuthenticationResult
}

func (server Server) Authenticate(writer http.ResponseWriter, request *http.Request) *AuthenticationResult {
	if server.Authenticator == nil {
		// No authenticator, everyone is anonymous
		return &AuthenticationResult{
			Role: AnonymousRole,
		}
	}

	return server.Authenticator.Authenticate(writer, request)
}
