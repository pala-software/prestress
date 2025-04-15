package oauth

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"gitlab.com/pala-software/prestress/pkg/prestress"
)

type OAuth struct {
	Disable          bool
	ClientId         string
	ClientSecret     string
	IntrospectionUrl *url.URL
}

// Construct OAuth Feature and read configuration from environment variables.
func OAuthFromEnv() *OAuth {
	var err error
	feature := OAuth{}

	feature.Disable = os.Getenv("PRESTRESS_AUTH_DISABLE") == "1"
	if feature.Disable {
		// No need to parse more configuration
		return &feature
	}

	feature.IntrospectionUrl, err = url.Parse(
		os.Getenv("PRESTRESS_OAUTH_INTROSPECTION_URL"),
	)
	if err != nil {
		panic(err)
	}
	if feature.IntrospectionUrl.String() == "" {
		panic("empty or unset PRESTRESS_OAUTH_INTROSPECTION_URL")
	}

	feature.ClientId = os.Getenv("PRESTRESS_OAUTH_CLIENT_ID")
	if feature.ClientId == "" {
		panic("empty or unset PRESTRESS_OAUTH_CLIENT_ID")
	}

	feature.ClientSecret = os.Getenv("PRESTRESS_OAUTH_CLIENT_SECRET")
	if feature.ClientSecret == "" {
		panic("empty or unset PRESTRESS_OAUTH_CLIENT_SECRET")
	}

	return &feature
}

func (feature OAuth) Apply(server *prestress.Server) error {
	if !feature.Disable {
		server.Authenticator = feature
	}
	return nil
}

func (feature OAuth) Authenticate(
	writer http.ResponseWriter,
	request *http.Request,
) *prestress.AuthenticationResult {
	authorization := request.Header.Get("Authorization")
	if authorization == "" {
		return &prestress.AuthenticationResult{
			Role: prestress.AnonymousRole,
		}
	}

	if !strings.HasPrefix(authorization, "Bearer ") {
		writer.WriteHeader(401)
		return nil
	}

	introspectionUrl := *feature.IntrospectionUrl
	introspectionUrl.User = url.UserPassword(
		feature.ClientId,
		feature.ClientSecret,
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

	var introspection map[string]any
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
		role = prestress.AuthenticatedRole
	}

	return &prestress.AuthenticationResult{
		Variables: introspection,
		Role:      role,
	}
}
