package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const anonymousRole = "anonymous"

type introspectionResponse struct {
	Active  bool   `json:"active"`
	Subject string `json:"sub"`
	Role    string `json:"role"`
}

type authenticationResult struct {
	// Maybe be empty if user is not recognized (anonymous).
	UserId string

	// Always set to some role name. It's anonymous if user is not authenticated.
	RoleName string
}

// If user can be authenticated from the request, pointer to result is returned.
// Otherwise response is written and nil is returned.
func (server Server) authenticate(writer http.ResponseWriter, request *http.Request) *authenticationResult {
	authorization := request.Header.Get("Authorization")
	if server.disableAuth || authorization == "" {
		return &authenticationResult{
			UserId:   "",
			RoleName: anonymousRole,
		}
	}

	if !strings.HasPrefix(authorization, "Bearer ") {
		writer.WriteHeader(500)
		return nil
	}

	introspectionUrl := *server.introspectionUrl
	introspectionUrl.RawQuery = url.Values{
		"token": []string{strings.TrimPrefix(authorization, "Bearer ")},
	}.Encode()
	introspectionUrl.User = url.UserPassword(server.clientId, server.clientSecret)
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

	introspection := introspectionResponse{
		Active: false,
		Role:   "",
	}
	err = json.Unmarshal(body, &introspection)
	if err != nil {
		fmt.Println(err)
		writer.WriteHeader(500)
		return nil
	}

	if !introspection.Active ||
		introspection.Subject == "" ||
		introspection.Role == "" {
		writer.WriteHeader(401)
		return nil
	}

	return &authenticationResult{
		UserId:   introspection.Subject,
		RoleName: introspection.Role,
	}
}
