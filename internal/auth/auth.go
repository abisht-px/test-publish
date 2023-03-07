package auth

import (
	"context"
	"fmt"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

func GetBearerToken(ctx context.Context, issuerURL, clientID, clientSecret, username, password string) (string, error) {
	provider, err := oidc.NewProvider(ctx, issuerURL)
	if err != nil {
		return "", fmt.Errorf("instantiating OIDC provider for %q: %w", issuerURL, err)
	}
	config := oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     provider.Endpoint(),
	}
	token, err := config.PasswordCredentialsToken(ctx, username, password)
	if err != nil {
		return "", fmt.Errorf("getting token from endpoint %q: %w", config.Endpoint.TokenURL, err)
	}
	return token.AccessToken, nil
}
