package eventbridge

import (
	"os"
	"time"

	"github.com/fogfish/swarm"
	"github.com/fogfish/swarm/kernel/encoding"
)

type Option func(*Client)

var defs = []Option{WithConfig(), WithEnv()}

func WithConfig(opts ...swarm.Option) Option {
	return func(c *Client) {
		config := swarm.NewConfig()
		for _, opt := range opts {
			opt(&config)
		}

		// Mandatory overrides
		config.PollFrequency = 5 * time.Microsecond
		config.Codec = encoding.NewCodecPacket()

		c.config = config
	}
}

func WithService(service EventBridge) Option {
	return func(c *Client) {
		c.service = service
	}
}

const EnvEventBus = "CONFIG_SWARM_EVENT_BUS"

func WithEnv() Option {
	return func(c *Client) {
		if val, has := os.LookupEnv(EnvEventBus); has {
			c.bus = val
		}
	}
}
