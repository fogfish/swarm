package websocket

import (
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/fogfish/swarm"
)

// Helper function to fetch principal identity from the context of websocket message
func PrincipalOf[T any](msg swarm.Msg[T]) (string, error) {
	ctx, ok := msg.IOContext.(*events.APIGatewayWebsocketProxyRequestContext)
	if !ok {
		return "", fmt.Errorf("invalid message context: %T", msg.IOContext)
	}

	// Note: Authorizer context is defined by authorizer function (see auth lambda).
	//       The context value depends on the auth type e.g. basic, jwt, etc).
	//       There are mandatory fields exists in any context: `auth`, `principalId`.
	//       In case of `jwt` the following fields exists: `iss`, `sub`, `exp`, `nbf`, `iat`, `scope`.
	//       In case of `basic` the fields `sub`, `scope` is simulated.
	auth, ok := ctx.Authorizer.(map[string]any)
	if !ok {
		return "", fmt.Errorf("invalid authorizer context: %T", ctx.Authorizer)
	}

	value, has := auth["principalId"]
	if !has {
		return "", fmt.Errorf("unknown principal in the authorizer context")
	}

	principal, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("invalid principal type: %T", value)
	}

	return principal, nil
}
