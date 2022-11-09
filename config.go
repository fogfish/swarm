package swarm

import "time"

type Policy int

const (
	PolicyAtMostOnce Policy = iota
	PolicyAtLeastOnce
	PolicyExactlyOnce
)

type Config struct {
	// Instance of AWS Service, ...
	Service any

	// Quality of Service Policy
	Policy Policy

	// Queue capacity
	EnqueueCapacity int
	DequeueCapacity int

	// Agent is a direct performer of the event.
	// A software service that emits action to the stream.
	Agent string

	// Frequency to poll broker api
	PollFrequency time.Duration

	// Time To Flight is a time required by the client to acknowledge the message
	TimeToFlight time.Duration

	// Timeout for any network operations
	NetworkTimeout time.Duration
}

func NewConfig() *Config {
	return &Config{
		Policy:          PolicyAtLeastOnce,
		EnqueueCapacity: 0,
		DequeueCapacity: 0,
		Agent:           "github.com/fogfish/swarm",
		PollFrequency:   10 * time.Millisecond,
		TimeToFlight:    5 * time.Second,
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

// ...
func WithNetworkTimeout(t time.Duration) Option {
	return func(conf *Config) {
		conf.NetworkTimeout = t
	}
}

// AtMostOnce is best effort policy, where a message is published without any
// formal acknowledgement of receipt, and it isn't replayed.
//
// The policy only impacts behavior of Golang channels created by the broker
func WithPolicyAtMostOnce(n int) Option {
	return func(conf *Config) {
		conf.Policy = PolicyAtMostOnce
		conf.EnqueueCapacity = n
		conf.DequeueCapacity = n
	}
}

// AtLeastOnce policy ensures delivery of the message to broker
//
// The policy only impacts behavior of Golang channels created by the broker
func WithPolicyAtLeastOnce(n int) Option {
	return func(conf *Config) {
		conf.Policy = PolicyAtLeastOnce
		conf.EnqueueCapacity = 0
		conf.DequeueCapacity = n
	}
}
