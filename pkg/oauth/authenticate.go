package oauth

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"

	"gitlab.com/pala-software/prestress/pkg/auth"
)

func (oauth OAuth) Authenticate(
	request *http.Request,
) (res *auth.AuthenticationResult, err error) {
	authorization := request.Header.Get("Authorization")
	if authorization == "" {
		res = &auth.AuthenticationResult{
			Role: auth.AnonymousRole,
		}
		return
	}

	if !strings.HasPrefix(authorization, "Bearer ") {
		err = auth.ErrAuthenticationFailed
		return
	}

	introspectionUrl := *oauth.IntrospectionUrl
	introspectionUrl.User = url.UserPassword(
		oauth.ClientId,
		oauth.ClientSecret,
	)
	requestBody := strings.NewReader(url.Values{
		"token": []string{strings.TrimPrefix(authorization, "Bearer ")},
	}.Encode())
	response, err := http.Post(
		introspectionUrl.String(),
		"application/x-www-form-urlencoded",
		requestBody,
	)
	if err != nil {
		return
	}
	if response.StatusCode != 200 {
		err = auth.ErrAuthenticationFailed
		return
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return
	}

	var introspection map[string]any
	err = json.Unmarshal(body, &introspection)
	if err != nil {
		return
	}

	active, ok := introspection["active"].(bool)
	if !ok {
		err = errors.New("unexpected type of 'active' property on token")
		return
	}

	if !active {
		err = auth.ErrAuthenticationFailed
		return
	}

	delete(introspection, "active")

	role := ""
	if _, exists := introspection["role"]; exists {
		role, ok = introspection["role"].(string)
		if !ok {
			err = errors.New("unexpected type of 'role' property on token")
			return
		}
	}

	if role == "" {
		role = auth.AuthenticatedRole
	}

	res = &auth.AuthenticationResult{
		Variables: introspection,
		Role:      role,
	}
	return
}
