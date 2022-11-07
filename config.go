package swarm

import "time"

type Config struct {
	// Instance of AWS Service
	Service any

	// Agent is a direct performer of the event.
	// A software service that emits action to the stream.
	Agent string

	// Frequency to poll broker api
	PollFrequency time.Duration

	// Time To Flight ...
	TimeToFlight time.Duration
}

func NewConfig() *Config {
	return &Config{
		Agent:         "github.com/fogfish/swarm",
		PollFrequency: 10 * time.Millisecond,
		TimeToFlight:  5 * time.Second,
	}
}

type Option func(conf *Config)

// Configure AWS Service for broker instance
func WithService(service any) Option {
	return func(conf *Config) {
		conf.Service = service
	}
}

// Agent is a direct performer of the event.
// A software service that emits action to the stream.
func WithAgent(agent string) Option {
	return func(conf *Config) {
		conf.Agent = agent
	}
}

// Frequency to poll broker api
func WithPollFrequency(t time.Duration) Option {
	return func(conf *Config) {
		conf.PollFrequency = t
	}
}

// ...
func WithTimeToFlight(t time.Duration) Option {
	return func(conf *Config) {
		conf.TimeToFlight = t
	}
}

// AtMostOnce is best effort policy, where a message is published without any
// formal acknowledgement of receipt, and isn't replayed.
//
// The policy only impacts behavior of Golang channels created by the broker
func WithPolicyAtMostOnce(n int) Option {
	return func(conf *Config) {

	}
}

// AtLeastOnce policy ensures delivery of the message to broker
//
// The policy only impacts behavior of Golang channels created by the broker
func WithPolicyAtLeastOnce() Option {
	return func(conf *Config) {

	}
}
