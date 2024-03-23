package main

import (
	"context"
	"errors"
	"log/slog"
	"net/url"
	"os"
	"time"

	"github.com/auth0/go-jwt-middleware/v2/jwks"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

var (
	none         = events.APIGatewayCustomAuthorizerResponse{}
	errForbidden = errors.New("forbidden")
)

type Claims struct {
	Scope string `json:"scope"`
}

func (c Claims) Validate(ctx context.Context) error { return nil }

func main() {
	uri := os.Getenv("CONFIG_SWARM_WS_AUTHORIZER_ISS")
	issuer, err := url.Parse(uri)
	if err != nil {
		slog.Error("Invalid issuer url.", "err", err, "url", uri)
		panic(err)
	}

	provider := jwks.NewCachingProvider(issuer, 5*time.Minute)

	auth, err := validator.New(
		provider.KeyFunc,
		validator.RS256,
		uri,
		[]string{os.Getenv("CONFIG_SWARM_WS_AUTHORIZER_AUD")},
		validator.WithCustomClaims(func() validator.CustomClaims { return &Claims{} }),
		validator.WithAllowedClockSkew(time.Minute),
	)
	if err != nil {
		slog.Error("Validator failed.", "err", err)
		panic(err)
	}

	lambda.Start(
		func(evt events.APIGatewayV2CustomAuthorizerV1Request) (events.APIGatewayCustomAuthorizerResponse, error) {
			token, has := evt.QueryStringParameters["token"]
			if !has {
				return none, errForbidden
			}

			claims, err := auth.ValidateToken(context.Background(), token)
			if err != nil {
				return none, errForbidden
			}

			switch c := claims.(type) {
			case *validator.ValidatedClaims:
				ctx := map[string]any{
					"iss": c.RegisteredClaims.Issuer,
					"sub": c.RegisteredClaims.Subject,
					// "aud": c.RegisteredClaims.Audience,
					"exp":   c.RegisteredClaims.Expiry,
					"nbf":   c.RegisteredClaims.NotBefore,
					"iat":   c.RegisteredClaims.IssuedAt,
					"scope": c.CustomClaims.(*Claims).Scope,
				}

				return accessPolicy(c.RegisteredClaims.Subject, evt.MethodArn, ctx), nil
			default:
				return none, errForbidden
			}
		},
	)
}

func accessPolicy(principal, method string, context map[string]any) events.APIGatewayCustomAuthorizerResponse {
	return events.APIGatewayCustomAuthorizerResponse{
		PrincipalID: principal,
		PolicyDocument: events.APIGatewayCustomAuthorizerPolicy{
			Version: "2012-10-17",
			Statement: []events.IAMPolicyStatement{
				{
					Action:   []string{"execute-api:*"},
					Effect:   "Allow",
					Resource: []string{method},
				},
			},
		},
		Context: context,
	}
}
