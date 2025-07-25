package oauth

import (
	"net/url"
	"os"
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

func (feature OAuth) Provider() any {
	return feature.Register
}

func (feature *OAuth) Register() *OAuth {
	return feature
}
