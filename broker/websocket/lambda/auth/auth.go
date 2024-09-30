//
// Copyright (C) 2021 - 2022 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package main

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"log/slog"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/auth0/go-jwt-middleware/v2/jwks"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	_ "github.com/fogfish/logger/v3"
)

func main() {
	basic, err := NewAuthBasic()
	if err != nil {
		slog.Warn("Basic Auth disabled.")
		basic = nil
	}

	jwt, err := NewAuthJWT()
	if err != nil {
		slog.Warn("JWT Auth disabled.")
		jwt = nil
	}

	lambda.Start(
		func(evt events.APIGatewayV2CustomAuthorizerV1Request) (events.APIGatewayCustomAuthorizerResponse, error) {
			tkn, has := evt.QueryStringParameters["apikey"]
			if !has || len(tkn) == 0 {
				return None, ErrForbidden
			}

			if jwt != nil && strings.HasPrefix(tkn, "ey") {
				principal, context, err := jwt.Validate(tkn)
				if err != nil {
					return None, ErrForbidden
				}

				return AccessPolicy(principal, evt.MethodArn, context), nil
			}

			if basic != nil {
				principal, context, err := basic.Validate(tkn)
				if err != nil {
					return None, ErrForbidden
				}

				return AccessPolicy(principal, evt.MethodArn, context), nil
			}

			return None, ErrForbidden
		},
	)

}

var (
	None         = events.APIGatewayCustomAuthorizerResponse{}
	ErrForbidden = errors.New("forbidden")
)

//------------------------------------------------------------------------------

// Grant the access to WebSocket with the policy
func AccessPolicy(principal, method string, context map[string]any) events.APIGatewayCustomAuthorizerResponse {
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

//------------------------------------------------------------------------------

type AuthBasic struct{ access, secret string }

func NewAuthBasic() (*AuthBasic, error) {
	access := os.Getenv("CONFIG_SWARM_WS_AUTHORIZER_ACCESS")
	secret := os.Getenv("CONFIG_SWARM_WS_AUTHORIZER_SECRET")

	if access == "" || secret == "" {
		return nil, errors.New("basic auth is not configured")
	}

	return &AuthBasic{
		access: access,
		secret: secret,
	}, nil
}

func (auth *AuthBasic) Validate(apikey string) (string, map[string]any, error) {
	c, err := base64.RawStdEncoding.DecodeString(apikey)
	if err != nil {
		return "", nil, ErrForbidden
	}

	access, secret, ok := strings.Cut(string(c), ":")
	if !ok {
		return "", nil, ErrForbidden
	}

	gaccess := sha256.Sum256([]byte(access))
	gsecret := sha256.Sum256([]byte(secret))
	haccess := sha256.Sum256([]byte(auth.access))
	hsecret := sha256.Sum256([]byte(auth.secret))

	accessMatch := (subtle.ConstantTimeCompare(gaccess[:], haccess[:]) == 1)
	secretMatch := (subtle.ConstantTimeCompare(gsecret[:], hsecret[:]) == 1)

	if accessMatch && secretMatch {
		return access, map[string]any{"auth": "basic"}, nil
	}

	return "", nil, ErrForbidden
}

//------------------------------------------------------------------------------

type AuthJWT struct {
	*validator.Validator
}

type Claims struct {
	Scope string `json:"scope"`
}

func (c Claims) Validate(ctx context.Context) error { return nil }

func NewAuthJWT() (*AuthJWT, error) {
	iss := os.Getenv("CONFIG_SWARM_WS_AUTHORIZER_ISS")
	aud := os.Getenv("CONFIG_SWARM_WS_AUTHORIZER_AUD")

	if iss == "" || aud == "" {
		return nil, errors.New("jwt auth is not configured")
	}

	issuer, err := url.Parse(iss)
	if err != nil {
		return nil, err
	}

	provider := jwks.NewCachingProvider(issuer, 5*time.Minute)

	auth, err := validator.New(
		provider.KeyFunc,
		validator.RS256,
		iss,
		[]string{aud},
		validator.WithCustomClaims(func() validator.CustomClaims { return &Claims{} }),
		validator.WithAllowedClockSkew(time.Minute),
	)
	if err != nil {
		return nil, err
	}

	return &AuthJWT{Validator: auth}, nil
}

func (auth *AuthJWT) Validate(token string) (string, map[string]any, error) {
	claims, err := auth.ValidateToken(context.Background(), token)
	if err != nil {
		return "", nil, ErrForbidden
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
			"auth":  "jwt",
		}

		return c.RegisteredClaims.Subject, ctx, nil
	default:
		return "", nil, ErrForbidden
	}
}
