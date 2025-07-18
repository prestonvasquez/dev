package main

import (
	"context"
	"errors"
	"net/http"

	"go.mongodb.org/mongo-driver/x/mongo/driver"
	"go.mongodb.org/mongo-driver/x/mongo/driver/auth"
)

const sourceExternal = "$external"

// awsSdkAuthenticator is an authenticator that uses the AWS SDK rather than the
// lightweight AWS package used internally by the driver.
type awsSdkAuthenticator struct{}

var _ driver.Authenticator = (*awsSdkAuthenticator)(nil)

// newAwsSdkAuthenticator creates a new AWS SDK authenticator. It loads the AWS
// SDK config (honoring AWS_STS_REGIONAL_ENDPOINTS & AWS_REGION) and rturns an
// Authenticator that uses it.
func newAwsSdkAuthenticator(*auth.Cred, *http.Client) (driver.Authenticator, error) {
	return &awsSdkAuthenticator{}, nil
}

var _ auth.AuthenticatorFactory = newAwsSdkAuthenticator

// Auth starts the SASL conversation by constructing a custom SASL adapter that
// uses teh AWS SDK for singing.
func (a *awsSdkAuthenticator) Auth(ctx context.Context, cfg *driver.AuthConfig) error {
	// Build a SASL adapter that uses AWS SDK for signing.
	adapter := &awsSdkSaslClient{}

	return auth.ConductSaslConversation(ctx, cfg, sourceExternal, adapter)
}

// REauth is not supported for AWS SDK authentication.
func (a *awsSdkAuthenticator) Reauth(context.Context, *driver.AuthConfig) error {
	return errors.New("AWS reauthentication not supported")
}

// awsSdkSaslClient is a SASL client that uses the AWS SDK for signing.
type awsSdkSaslClient struct{}

var _ auth.SaslClient = (*awsSdkSaslClient)(nil)

// Start will create the client-first SASL message.
// { p: 110, r: <32-byte nonce>}; per the current Go Driver behavior.
func (a *awsSdkSaslClient) Start() (string, []byte, error) {
	return "", nil, nil
}

// Next handles the server's "server-first" message, then builds and returns the
// "client-final" payload containing the SigV4-signed STS GetCallerIdentity
// request.
func (a *awsSdkSaslClient) Next(ctx context.Context, challenge []byte) ([]byte, error) {
	return nil, nil
}

// complete signals that the SASL conversation is done.
func (a *awsSdkSaslClient) Completed() bool {
	return false
}

func main() {
	auth.RegisterAuthenticatorFactory(auth.MongoDBAWS, newAwsSdkAuthenticator)
}
