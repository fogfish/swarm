//
// Copyright (C) 2021 - 2022 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package main

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

var (
	none         = events.APIGatewayCustomAuthorizerResponse{}
	errForbidden = errors.New("forbidden")
)

// Inspired by https://gist.github.com/praveen001/1b045d1c31cd9c72e4e6638e9f883f83

func main() {
	lambda.Start(
		func(evt events.APIGatewayV2CustomAuthorizerV1Request) (events.APIGatewayCustomAuthorizerResponse, error) {
			key, has := evt.QueryStringParameters["apikey"]
			if !has {
				return none, errForbidden
			}

			if principal, context, err := validate(key); err == nil {
				return accessPolicy(principal, evt.MethodArn, context), nil
			}
			return none, errForbidden

		},
	)
}

func validate(apikey string) (string, map[string]any, error) {
	c, err := base64.RawStdEncoding.DecodeString(apikey)
	if err != nil {
		return "", nil, errForbidden
	}

	access, secret, ok := strings.Cut(string(c), ":")
	if !ok {
		return "", nil, errForbidden
	}

	gaccess := sha256.Sum256([]byte(access))
	gsecret := sha256.Sum256([]byte(secret))
	haccess := sha256.Sum256([]byte(os.Getenv("CONFIG_SWARM_WS_AUTHORIZER_ACCESS")))
	hsecret := sha256.Sum256([]byte(os.Getenv("CONFIG_SWARM_WS_AUTHORIZER_SECRET")))

	accessMatch := (subtle.ConstantTimeCompare(gaccess[:], haccess[:]) == 1)
	secretMatch := (subtle.ConstantTimeCompare(gsecret[:], hsecret[:]) == 1)

	if accessMatch && secretMatch {
		return access, map[string]any{}, nil
	}

	return "", nil, errForbidden
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
