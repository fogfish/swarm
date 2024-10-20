package websocket

import (
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/fogfish/it/v2"
	"github.com/fogfish/swarm"
)

func TestPrincipalOf(t *testing.T) {
	t.Run("IOContext-None", func(t *testing.T) {
		msg := swarm.Msg[string]{}
		it.Then(t).Should(
			it.Error(PrincipalOf(msg)).Contain("invalid message context"),
		)
	})

	t.Run("IOContext-Bad", func(t *testing.T) {
		msg := swarm.Msg[string]{IOContext: "context"}
		it.Then(t).Should(
			it.Error(PrincipalOf(msg)).Contain("invalid message context"),
		)
	})

	t.Run("AuthContext-Bad", func(t *testing.T) {
		msg := swarm.Msg[string]{IOContext: &events.APIGatewayWebsocketProxyRequestContext{
			Authorizer: "authroizer",
		}}
		it.Then(t).Should(
			it.Error(PrincipalOf(msg)).Contain("invalid authorizer context"),
		)
	})

	t.Run("Principal-None", func(t *testing.T) {
		msg := swarm.Msg[string]{IOContext: &events.APIGatewayWebsocketProxyRequestContext{
			Authorizer: map[string]any{},
		}}
		it.Then(t).Should(
			it.Error(PrincipalOf(msg)).Contain("unknown principal in the authorizer context"),
		)
	})

	t.Run("Principal-Bad", func(t *testing.T) {
		msg := swarm.Msg[string]{IOContext: &events.APIGatewayWebsocketProxyRequestContext{
			Authorizer: map[string]any{
				"principalId": 1,
			},
		}}
		it.Then(t).Should(
			it.Error(PrincipalOf(msg)).Contain("invalid principal type"),
		)
	})

	t.Run("Principal", func(t *testing.T) {
		msg := swarm.Msg[string]{IOContext: &events.APIGatewayWebsocketProxyRequestContext{
			Authorizer: map[string]any{
				"principalId": "test",
			},
		}}
		val, err := PrincipalOf(msg)
		it.Then(t).Should(
			it.Nil(err),
			it.Equal(val, "test"),
		)
	})
}
