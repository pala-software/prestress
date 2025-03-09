package prestress

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const anonymousRole = "anonymous"
const authenticatedRole = "authenticated"

type AuthenticationResult struct {
	// Maybe be empty if user is not recognized (anonymous).
	Variables map[string]interface{}

	// Always set to some role name. It's anonymous if user is not authenticated.
	Role string
}

// If user can be authenticated from the request, pointer to result is returned.
// Otherwise response is written and nil is returned.
// TODO: Test
func (server Server) Authenticate(writer http.ResponseWriter, request *http.Request) *AuthenticationResult {
	authorization := request.Header.Get("Authorization")
	if server.DisableAuth || authorization == "" {
		return &AuthenticationResult{
			Variables: map[string]interface{}{},
			Role:      anonymousRole,
		}
	}

	if !strings.HasPrefix(authorization, "Bearer ") {
		writer.WriteHeader(500)
		return nil
	}

	introspectionUrl := *server.IntrospectionUrl
	introspectionUrl.RawQuery = url.Values{
		"token": []string{strings.TrimPrefix(authorization, "Bearer ")},
	}.Encode()
	introspectionUrl.User = url.UserPassword(server.ClientId, server.ClientSecret)
	response, err := http.Get(introspectionUrl.String())
	if err != nil {
		fmt.Println(err)
		writer.WriteHeader(500)
		return nil
	}
	if response.StatusCode != 200 {
		writer.WriteHeader(401)
		return nil
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Println(err)
		writer.WriteHeader(500)
		return nil
	}

	introspection := map[string]interface{}{}
	err = json.Unmarshal(body, &introspection)
	if err != nil {
		fmt.Println(err)
		writer.WriteHeader(500)
		return nil
	}

	active, ok := introspection["active"].(bool)
	if !ok {
		fmt.Println("unexpected type of 'active' property on token")
		writer.WriteHeader(500)
		return nil
	}

	if !active {
		writer.WriteHeader(401)
		return nil
	}

	delete(introspection, "active")

	role := ""
	if _, exists := introspection["role"]; exists {
		role, ok = introspection["role"].(string)
		if !ok {
			fmt.Println("unexpected type of 'role' property on token")
			writer.WriteHeader(500)
			return nil
		}
	}

	if role == "" {
		role = authenticatedRole
	}

	return &AuthenticationResult{
		Variables: introspection,
		Role:      role,
	}
}
