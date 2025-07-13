# ADR-001: Configuration API Redesign

**Status**: Accepted  
**Date**: 2025-01-07  
**Deciders**: @fogfish  
**Technical Story**: Configuration Overhead and Poor UX  

## Context

The current swarm library v0.22.x configuration system suffers from significant usability issues:

1. **Configuration Overhead**: Users face overwhelming choice with 80% kernel options + 20% broker options
2. **Nested Configuration**: The `WithConfig()` pattern creates unintuitive nesting that obscures the distinction between broker and kernel configuration
4. **Cognitive Load**: Users must understand internal kernel vs broker architecture to configure the library

### Current Problem Example

```go
// Current: Poor UX due to overwhelming options and nesting
q, err := sqs.New("queue-name",
  sqs.WithBatchSize(5),           // Broker-specific - intuitive
  sqs.WithService(mockSQS),       // Broker-specific - intuitive
  sqs.WithConfig(                 // NESTING PROBLEM - not intuitive!
    swarm.WithSource("my-service"),
    swarm.WithRetry(backoff.Exp(10*time.Millisecond, 10, 0.5)),
    swarm.WithPollFrequency(10*time.Second),
    swarm.WithTimeToFlight(30*time.Second),
    swarm.WithNetworkTimeout(10*time.Second),
    swarm.WithLogStdErr(),
    swarm.WithConfigFromEnv(),
    swarm.WithPolicyAtLeastOnce(100),
  ),
)
```

This forces users to:
- Understand the 80/20 split between kernel and broker options
- Remember which options go where
- Navigate nested configuration structure
- Make decisions about kernel options they rarely need to change

## Decision

We will implement a **Hybrid Configuration Pattern** that uses:

1. **Builder Pattern for Brokers**: Simple, discoverable API for broker-specific configuration
2. **opts Pattern for Kernel**: Bundle kernel complexity into single `WithKernel()` option
3. **Hide Kernel Concept**: Users shouldn't need to understand internal architecture
4. **Sensible Defaults**: 95% of users should not need to configure kernel options. Builder must work with zero configuration.
5. **Breaking Change**: This is a major API change - no backward compatibility required (v0.23.x to be released)


### New API Design

```go
// Simple case (95% of users) - ultra-clean with sensible defaults
client := sqs.Endpoint().Build("queue")
sender := sqs.Emitter().Build("queue")  
receiver := sqs.Listener().Build("queue")

// Broker customization - still simple
processor := sqs.Endpoint().
  WithBatchSize(10).
  WithService(customSQS).
  Build("work-queue")

// Power user case - kernel configuration when needed  
advanced := sqs.Endpoint().
  WithKernel(
    swarm.WithSource("advanced-service"),
    swarm.WithRetry(backoff.Exp(50*time.Millisecond, 15, 0.7)),
    swarm.WithPollFrequency(500*time.Millisecond),
    swarm.WithTimeToFlight(2*time.Minute),
    swarm.WithLogStdErr(),
  ).
  WithBatchSize(25).
  WithService(customSQS).
  Build("complex-queue")
```

## Architecture

### Core Components

1. **Builder**: There are few entry point for broker configuration
   ```go
   func Endpoint() *Builder
   func Listener() *Builder
   func Emitter() *Builder
   ```

2. **Builder Methods**: Fluent API for broker-specific options
   ```go
   func (b *Builder) WithBatchSize(size int) *Builder
   func (b *Builder) WithService(svc SQS) *Builder
   ```

3. **Kernel Configuration**: Single bundled option for advanced users
   ```go
   func (b *Builder) WithKernel(opts ...opts.Option[swarm.Config]) *Builder
   ```

4. **Terminal Functions**: Create the actual communication channels
   ```go
   func (b *Builder) Build(queue string) (*kernel.Enqueuer, error)
   ```

### Naming Conventions

- **Entry Point**: - clear single-purpose builder intent
  -  `Endpoint()` - emphasizes duplex Emitter/Listener channels for communication;
  -  `Listener()` - emphasizes receiver only channels;
  -  `Emitter()` - emphasizes sender only channels;
  -  core library abstraction (Go channels for distributed communication)
- **Queue Parameter**: Always at the end as mandatory parameter: `Build("queue-name")`

### Cross-Broker Consistency

All brokers follow the same pattern:

```go
// SQS
sqs.Endpoint().WithKernel(...).WithBatchSize(5).NewClient("queue")

// EventBridge  
eventbridge.Endpoint().WithKernel(...).WithEventBus("bus").NewClient()

// WebSocket
websocket.Endpoint().WithKernel(...).WithService(gateway).NewClient()
```

## Implementation Guidelines

### For Human Developers

1. **Implementation Order**: Start with SQS broker as reference implementation
2. **Builder Structure**: Each broker implements ~50 lines of builder code
5. **Documentation**: Update all examples to use new pattern

### For AI Agents

When implementing this pattern:

```go
// Template for broker builder implementation
type EmitterBuilder struct { *builder[*EmitterBuilder] }

func Emitter() *EmitterBuilder {
  b := &EmitterBuilder{}
	b.builder = newBuilder(b)
	return b
}

func (b *EmitterBuilder) Build(bus string) (*kernel.EmitterCore, error) {
	client, err := b.build(bus)
	if err != nil {
		return nil, err
	}
	return kernel.NewEmitter(client, client.config), nil
}

type builder[T any] struct {
	b          T
	kernelOpts []opts.Option[swarm.Config]
	service    EventBridge
}

// newBuilder creates new builder for EventBridge broker configuration.
func newBuilder[T any](b T) *builder[T] {
	kopts := []opts.Option[swarm.Config]{
		swarm.WithLogStdErr(),
		swarm.WithConfigFromEnv(),
	}
	if val := os.Getenv(EnvConfigSourceEventBridge); val != "" {
		kopts = append(kopts, swarm.WithAgent(val))
	}

	return &builder[T]{
		b:          b,
		kernelOpts: kopts,
	}
}

// WithKernel configures swarm kernel options for advanced usage.
func (b *builder[T]) WithKernel(opts ...opts.Option[swarm.Config]) T {
	b.kernelOpts = append(b.kernelOpts, opts...)
	return b.b
}

// WithService configures AWS EventBridge client instance
func (b *builder[T]) WithService(service EventBridge) T {
	b.service = service
	return b.b
}

func (b *Builder) build(queue string) (*Client, error) {
	client := &Client{
		config:  swarm.NewConfig(),
		bus:     bus,
		service: b.service,
	}

	if err := opts.Apply(&client.config, b.kernelOpts); err != nil {
		return nil, err
	}

  // 4. Perform broker-specific initialization
}
```

## Constraints and Requirements

### Must Use
- `github.com/fogfish/opts` for kernel configuration - no new option abstractions
- Builder pattern for broker configuration - proven, IDE-friendly pattern

### Must Not
- Create boilerplate re-export of kernel options per broker (Alternative 1 rejected)
- Maintain backward compatibility (breaking change accepted)
- Expose kernel concept to end users

### Must Achieve
- 95% of users need zero kernel configuration
- Simple cases stay simple: `sqs.Endpoint().Build("queue")`
- Complex cases remain powerful when needed
- Consistent API across all brokers

## Consequences

### Positive
- **Dramatic UX improvement**: Configuration overhead reduced from overwhelming to minimal
- **Progressive complexity**: Simple by default, powerful when needed
- **Conceptual clarity**: Clear separation between broker behavior and kernel configuration
- **Consistent patterns**: Same mental model across all brokers

### Negative
- **Breaking change**: All existing code must be updated
- **Implementation effort**: ~50 lines of builder code per broker
- **Learning curve**: Existing users must adapt to new pattern

### Neutral
- **Code volume**: Similar total code volume, better organized
- **Type safety**: Maintained through opts library integration

## Alternatives Considered

1. **Direct Kernel Option Re-Export**: Rejected due to boilerplate (12+ lines per option per broker)
2. **Configuration Presets**: Rejected as still requires understanding kernel/broker split
3. **Environment-First**: Rejected as doesn't solve programmatic configuration UX
4. **Flattened API per Broker**: Rejected due to excessive boilerplate
5. **Smart WithConfig**: Rejected as maintains nesting problem

## Examples

### Migration Path

```go
// Before (current)
q, err := sqs.New("queue",
  sqs.WithBatchSize(5),
  sqs.WithConfig(
    swarm.WithSource("service"),
    swarm.WithLogStdErr(),
  ),
)

// After (new pattern)
q, err := sqs.Endpoint().
  WithKernel(
    swarm.WithSource("service"),
    swarm.WithLogStdErr(),
  ).
  WithBatchSize(5).
  Build("queue")

// Most common case becomes:
q, err := sqs.Endpoint().Build("queue") // That's it!
```

### Cross-Broker Examples

```go
// EventBridge
events := eventbridge.Emitter().
  WithKernel(swarm.WithSource("event-service")).
  WithEventBus("production-events").
  Build()

// WebSocket  
socket := websocket.Endpoint().
  WithKernel(swarm.WithPollFrequency(100*time.Millisecond)).
  WithService(gateway).
  Build()

// DynamoDB Events (read-only)
stream := eventddb.Listener().
  WithKernel(swarm.WithSource("ddb-stream")).
  Build()
```

## Success Metrics

1. **User Experience**: 95% of usage requires no kernel configuration
2. **Consistency**: Same pattern works across all 6+ brokers
3. **Simplicity**: Common case reduces from 10+ lines to 1 line of configuration

