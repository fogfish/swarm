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

	"github.com/fogfish/opts"
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

type Retry interface{ Retry(f func() error) error }

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

var (
	// Source is a direct performer of the event.
	// A software service that emits action to the stream.
	WithSource = opts.ForName[Config, string]("Source")

	// Define I/O backoff strategy
	// * backoff.Const(t, n) retry operation for N times, with T wait time in between
	// * backoff.Linear(t, n) retry operation for N times, with linear increments by T on each step
	// * backoff.Exp(t, n, f) retry operation for N times, with exponential increments by T on each step
	// * backoff.Empty() no retry
	WithRetry = opts.ForType[Config, Retry]()

	// Configure broker to route global errors to channel
	WithStdErr = opts.ForType[Config, chan<- error]()

	// Number of poller in the system
	WithPollerPool = opts.ForName[Config, int]("PollerPool")

	// Frequency to poll broker api
	WithPollFrequency = opts.ForName[Config, time.Duration]("PollFrequency")

	// Time To Flight for message from broker API to consumer
	WithTimeToFlight = opts.ForName[Config, time.Duration]("TimeToFlight")

	// Timeout for Network I/O
	WithNetworkTimeout = opts.ForName[Config, time.Duration]("NetworkTimeout")

	// AtMostOnce is best effort policy, where a message is published without any
	// formal acknowledgement of receipt, and it isn't replayed.
	//
	// The policy only impacts behavior of Golang channels created by the broker
	WithPolicyAtMostOnce = opts.ForName("CapRcv",
		func(c *Config, n int) error {
			c.Policy = PolicyAtMostOnce
			c.CapOut = n
			c.CapDlq = n
			c.CapRcv = n
			c.CapAck = n
			return nil
		})

	// AtLeastOnce policy ensures delivery of the message to broker
	//
	// The policy only impacts behavior of Golang channels created by the broker
	WithPolicyAtLeastOnce = opts.ForName("CapRcv",
		func(c *Config, n int) error {
			c.Policy = PolicyAtLeastOnce
			c.CapOut = 0
			c.CapDlq = 0
			c.CapRcv = n
			c.CapAck = n
			return nil
		},
	)

	// Fail fast the message if category is not known to kernel.
	WithFailOnUnknownCategory = opts.ForName[Config, bool]("FailOnUnknownCategory")
)

// Configure broker to log standard errors
func WithLogStdErr() opts.Option[Config] {
	return opts.Type[Config](
		func(c *Config) error {
			err := make(chan error)

			go func() {
				var x error
				for x = range err {
					slog.Error("Broker failed", "error", x)
				}
			}()

			c.StdErr = err
			return nil
		},
	)
}

// Configure from Environment, (all timers in seconds)
// - CONFIG_SWARM_POLL_FREQUENCY
// - CONFIG_SWARM_TIME_TO_FLIGHT
// - CONFIG_SWARM_NETWORK_TIMEOUT
func WithConfigFromEnv() opts.Option[Config] {
	return opts.Type[Config](
		func(c *Config) error {
			c.PollFrequency = durationFromEnv(EnvConfigPollFrequency, c.PollFrequency)
			c.TimeToFlight = durationFromEnv(EnvConfigTimeToFlight, c.TimeToFlight)
			c.NetworkTimeout = durationFromEnv(EnvConfigNetworkTimeout, c.NetworkTimeout)
			return nil
		},
	)
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
