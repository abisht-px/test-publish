package auth

import (
	"context"
	"fmt"
	"net/http"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

func GetAuthenticatedClientByPassword(
	ctx context.Context,
	issuerURL, clientID, clientSecret, username, password string,
) (*http.Client, error) {
	token, config, err := createTokenWithConfiguration(ctx, issuerURL, clientID, clientSecret, username, password)
	if err != nil {
		return nil, err
	}

	client := config.Client(ctx, token)
	return client, nil
}

func GetAuthenticatedClientByToken(ctx context.Context, token string) *http.Client {
	return oauth2.NewClient(ctx, GetTokenSourceByToken(token))
}

func GetTokenSourceByPassword(
	ctx context.Context,
	issuerURL, clientID, clientSecret, username, password string,
) (oauth2.TokenSource, error) {
	token, config, err := createTokenWithConfiguration(ctx, issuerURL, clientID, clientSecret, username, password)
	if err != nil {
		return nil, err
	}

	tokenSource := config.TokenSource(ctx, token)
	return tokenSource, nil
}

func GetTokenSourceByToken(token string) oauth2.TokenSource {
	t := &oauth2.Token{AccessToken: token}
	return oauth2.StaticTokenSource(t)
}

func createTokenWithConfiguration(
	ctx context.Context,
	issuerURL, clientID, clientSecret, username, password string,
) (*oauth2.Token, *oauth2.Config, error) {
	provider, err := oidc.NewProvider(ctx, issuerURL)
	if err != nil {
		return nil, nil, fmt.Errorf("instantiating OIDC provider for %q: %w", issuerURL, err)
	}
	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     provider.Endpoint(),
	}
	token, err := config.PasswordCredentialsToken(ctx, username, password)
	if err != nil {
		return nil, nil, err
	}

	return token, config, nil
}
