//
// Copyright (C) 2021 - 2022 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package swarm

import (
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/fogfish/swarm/kernel/backoff"
)

// Environment variable to config kernel
const (
	EnvConfigPollFrequency  = "CONFIG_SWARM_POLL_FREQUENCY"
	EnvConfigTimeToFlight   = "CONFIG_SWARM_TIME_TO_FLIGHT"
	EnvConfigNetworkTimeout = "CONFIG_SWARM_NETWORK_TIMEOUT"
)

// Grade of Service Policy
type Policy int

const (
	PolicyAtMostOnce Policy = iota
	PolicyAtLeastOnce
	PolicyExactlyOnce
)

type Retry interface {
	Retry(f func() error) error
}

type Codec interface {
	Encode([]byte) ([]byte, error)
	Decode([]byte) ([]byte, error)
}

type Config struct {
	// Source is a direct performer of the event.
	// A software service that emits action to the stream.
	Source string

	// Quality of Service Policy
	Policy Policy

	// Queue capacity (enhance with individual capacities)
	CapOut int
	CapDlq int
	CapRcv int
	CapAck int

	// Retry Policy for service calls
	Backoff Retry

	// Standard Error I/O channel
	StdErr chan<- error

	// Size of poller pool in the system
	PollerPool int

	// Frequency to poll broker api
	PollFrequency time.Duration

	// Time To Flight is a time required by the client to acknowledge the message
	TimeToFlight time.Duration

	// Timeout for any network operations
	NetworkTimeout time.Duration

	// Fail fast the message if category is not known to kernel.
	FailOnUnknownCategory bool

	// PacketCodec for binary packets
	PacketCodec Codec
}

func NewConfig() Config {
	return Config{
		Source:                "github.com/fogfish/swarm",
		Policy:                PolicyAtLeastOnce,
		CapOut:                0,
		CapDlq:                0,
		CapRcv:                0,
		CapAck:                0,
		Backoff:               backoff.Exp(10*time.Millisecond, 10, 0.5),
		PollerPool:            1,
		PollFrequency:         10 * time.Millisecond,
		TimeToFlight:          5 * time.Second,
		NetworkTimeout:        5 * time.Second,
		FailOnUnknownCategory: false,
	}
}

// Configuration option for queueing broker
type Option func(conf *Config)

// Source is a direct performer of the event.
// A software service that emits action to the stream.
func WithSource(agent string) Option {
	return func(conf *Config) {
		conf.Source = agent
	}
}

// Retry operation for N times, with T wait time in between
func WithRetryConstant(t time.Duration, n int) Option {
	return func(conf *Config) {
		conf.Backoff = backoff.Const(t, n)
	}
}

// Retry operation for N times, with linear increments by T on each step
func WithRetryLinear(t time.Duration, n int) Option {
	return func(conf *Config) {
		conf.Backoff = backoff.Linear(t, n)
	}
}

// Retry operation for N times, with exponential increments by T on each step
func WithRetryExponential(t time.Duration, n int, f float64) Option {
	return func(conf *Config) {
		conf.Backoff = backoff.Exp(t, n, f)
	}
}

// No retires
func WithRetryNo() Option {
	return func(conf *Config) {
		conf.Backoff = backoff.Empty()
	}
}

// Custom retry policy
func WithRetry(backoff Retry) Option {
	return func(conf *Config) {
		conf.Backoff = backoff
	}
}

// Configure broker to route global errors to channel
func WithStdErr(stderr chan<- error) Option {
	return func(conf *Config) {
		conf.StdErr = stderr
	}
}

// Configure broker to log standard errors
func WithLogStdErr() Option {
	err := make(chan error)

	go func() {
		var x error
		for x = range err {
			slog.Error("Broker failed", "error", x)
		}
	}()

	return func(conf *Config) {
		conf.StdErr = err
	}
}

// Number of poller in the system
func WithPollerPool(n int) Option {
	return func(conf *Config) {
		conf.PollerPool = n
	}
}

// Frequency to poll broker api
func WithPollFrequency(t time.Duration) Option {
	return func(conf *Config) {
		conf.PollFrequency = t
	}
}

// Time To Flight for message from broker API to consumer
func WithTimeToFlight(t time.Duration) Option {
	return func(conf *Config) {
		conf.TimeToFlight = t
	}
}

// Timeout for Network I/O
func WithNetworkTimeout(t time.Duration) Option {
	return func(conf *Config) {
		conf.NetworkTimeout = t
	}
}

// Configure from Environment, (all timers in seconds)
// - CONFIG_SWARM_POLL_FREQUENCY
// - CONFIG_SWARM_TIME_TO_FLIGHT
// - CONFIG_SWARM_NETWORK_TIMEOUT
func WithConfigFromEnv() Option {
	return func(conf *Config) {
		conf.PollFrequency = durationFromEnv(EnvConfigPollFrequency, conf.PollFrequency)
		conf.TimeToFlight = durationFromEnv(EnvConfigTimeToFlight, conf.TimeToFlight)
		conf.NetworkTimeout = durationFromEnv(EnvConfigNetworkTimeout, conf.NetworkTimeout)
	}
}

func durationFromEnv(key string, def time.Duration) time.Duration {
	val, has := os.LookupEnv(key)
	if !has {
		return def
	}

	sec, err := strconv.Atoi(val)
	if err != nil {
		return def
	}

	return time.Duration(sec) * time.Second
}

// AtMostOnce is best effort policy, where a message is published without any
// formal acknowledgement of receipt, and it isn't replayed.
//
// The policy only impacts behavior of Golang channels created by the broker
func WithPolicyAtMostOnce(n int) Option {
	return func(conf *Config) {
		conf.Policy = PolicyAtMostOnce
		conf.CapOut = n
		conf.CapDlq = n
		conf.CapRcv = n
		conf.CapAck = n
	}
}

// AtLeastOnce policy ensures delivery of the message to broker
//
// The policy only impacts behavior of Golang channels created by the broker
func WithPolicyAtLeastOnce(n int) Option {
	return func(conf *Config) {
		conf.Policy = PolicyAtLeastOnce
		conf.CapOut = 0
		conf.CapDlq = 0
		conf.CapRcv = n
		conf.CapAck = n
	}
}

// Configure codec for binary packets
func WithPacketCodec(codec Codec) Option {
	return func(conf *Config) {
		conf.PacketCodec = codec
	}
}

// Fail fast the message if category is not known to kernel.
func WithFailOnUnknownCategory() Option {
	return func(conf *Config) {
		conf.FailOnUnknownCategory = true
	}
}
