# AWS WebSocket API Broker

The sub-module implements swarm broker for AWS WebSocket API. See [the library documentation](../../README.md) for details.

Note the broker implements only infrastructure required for making serverless applications using WebSocket APIs.

Use cli https://github.com/vi/websocat for testing purposes.

```bash
websocat wss://0000000000.execute-api.eu-west-1.amazonaws.com/ws/\?apikey=dGVzdDp0ZXN0
{"action":"User", "id":"xxx", "text":"some text"}
```