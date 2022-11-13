//
// Copyright (C) 2021 - 2022 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the Apache License Version 2.0. See the LICENSE file for details.
// https://github.com/fogfish/swarm
//

package swarm

import (
	"time"

	"github.com/fogfish/swarm/internal/backoff"
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

type Config struct {
	// Instance of AWS Service, ...
	Service any

	// Source is a direct performer of the event.
	// A software service that emits action to the stream.
	Source string

	// Quality of Service Policy
	Policy Policy

	// Queue capacity
	EnqueueCapacity int
	DequeueCapacity int

	// Retry Policy for service calls
	Backoff Retry

	// Standard Error I/O channel
	StdErr chan<- error

	// Frequency to poll broker api
	PollFrequency time.Duration

	// Time To Flight is a time required by the client to acknowledge the message
	TimeToFlight time.Duration

	// Timeout for any network operations
	NetworkTimeout time.Duration
}

func NewConfig() Config {
	return Config{
		Source:          "github.com/fogfish/swarm",
		Policy:          PolicyAtLeastOnce,
		EnqueueCapacity: 0,
		DequeueCapacity: 0,
		Backoff:         backoff.Exp(10*time.Millisecond, 10, 0.5),
		PollFrequency:   10 * time.Millisecond,
		TimeToFlight:    5 * time.Second,
		NetworkTimeout:  5 * time.Second,
	}
}

// Configuration option for queueing broker
type Option func(conf *Config)

// Configure AWS Service for broker instance
func WithService(service any) Option {
	return func(conf *Config) {
		conf.Service = service
	}
}

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

// Configure Channel for global errors
func WithStdErr(stderr chan<- error) Option {
	return func(conf *Config) {
		conf.StdErr = stderr
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
