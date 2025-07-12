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
client := sqs.Channels().NewClient("queue")
sender := sqs.Channels().NewEnqueuer("queue")  
receiver := sqs.Channels().NewDequeuer("queue")

// Broker customization - still simple
processor := sqs.Channels().
  WithBatchSize(10).
  WithService(customSQS).
  NewClient("work-queue")

// Power user case - kernel configuration when needed  
advanced := sqs.Channels().
  WithKernel(
    swarm.WithSource("advanced-service"),
    swarm.WithRetry(backoff.Exp(50*time.Millisecond, 15, 0.7)),
    swarm.WithPollFrequency(500*time.Millisecond),
    swarm.WithTimeToFlight(2*time.Minute),
    swarm.WithLogStdErr(),
  ).
  WithBatchSize(25).
  WithService(customSQS).
  NewClient("complex-queue")
```

## Architecture

### Core Components

1. **Channel Builder**: Entry point for all broker configuration
   ```go
   func Channels() *Builder
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
   func (b *Builder) NewEnqueuer(queue string) (*kernel.Enqueuer, error)
   func (b *Builder) NewDequeuer(queue string) (*kernel.Dequeuer, error)  
   func (b *Builder) NewClient(queue string) (*kernel.Kernel, error)
   ```

### Naming Conventions

- **Entry Point**: `Channels()` - emphasizes core library abstraction (Go channels for distributed communication)
- **Specialized Functions**: `NewEnqueuer()`, `NewDequeuer()` - clear single-purpose intent
- **Full Capability**: `NewClient()` - intuitive bidirectional client pattern, hides kernel concept
- **Queue Parameter**: Always at the end as mandatory parameter: `New*("queue-name")`

### Cross-Broker Consistency

All brokers follow the same pattern:

```go
// SQS
sqs.Channels().WithKernel(...).WithBatchSize(5).NewClient("queue")

// EventBridge  
eventbridge.Channels().WithKernel(...).WithEventBus("bus").NewClient()

// WebSocket
websocket.Channels().WithKernel(...).WithService(gateway).NewClient()
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
type Builder struct {
  kernelConfig []opts.Option[swarm.Config]
  // broker-specific fields (e.g., batchSize, service)
}

func Channels() *Builder {
  return &Builder{}
}

func (b *Builder) WithKernel(opts ...opts.Option[swarm.Config]) *Builder {
  b.kernelOpts = append(b.kernelOpts, opts...)
  return b
}

// Broker-specific methods
func (b *Builder) WithBrokerOption(value Type) *Builder {
  b.brokerField = value
  return b
}

// Terminal functions
func (b *Builder) NewEnqueuer(queue string) (*kernel.Enqueuer, error) {
  client, err := b.build(queue)
  if err != nil {
    return nil, err
  }
  return kernel.NewEnqueuer(client, client.config), nil
}

func (b *Builder) NewDequeuer(queue string) (*kernel.Dequeuer, error) {
  client, err := b.build(queue)
  if err != nil {
    return nil, err
  }
  return kernel.NewDequeuer(client, client.config), nil
}

func (b *Builder) NewClient(queue string) (*kernel.Kernel, error) {
  client, err := b.build(queue)
  if err != nil {
    return nil, err
  }
  return kernel.New(
    kernel.NewEnqueuer(client, client.config),
    kernel.NewDequeuer(client, client.config),
  ), nil
}

func (b *Builder) build(queue string) (*Client, error) {
  kernelConfig := swarm.NewConfig()
  if err := opts.Apply(&kernelConfig, opt); err != nil {
    return nil, err
  }
  // 1. Create client with broker-specific fields
  // 3. Apply user kernel options (if any)
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
- Simple cases stay simple: `sqs.Channels().NewClient("queue")`
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
q, err := sqs.Channels().
  WithKernel(
    swarm.WithSource("service"),
    swarm.WithLogStdErr(),
  ).
  WithBatchSize(5).
  NewClient("queue")

// Most common case becomes:
q, err := sqs.Channels().NewClient("queue") // That's it!
```

### Cross-Broker Examples

```go
// EventBridge
events := eventbridge.Channels().
  WithKernel(swarm.WithSource("event-service")).
  WithEventBus("production-events").
  NewEnqueuer()

// WebSocket  
socket := websocket.Channels().
  WithKernel(swarm.WithPollFrequency(100*time.Millisecond)).
  WithService(gateway).
  NewClient()

// DynamoDB Events (read-only)
stream := eventddb.Channels().
  WithKernel(swarm.WithSource("ddb-stream")).
  NewDequeuer()
```

## Success Metrics

1. **User Experience**: 95% of usage requires no kernel configuration
2. **Consistency**: Same pattern works across all 6+ brokers
3. **Simplicity**: Common case reduces from 10+ lines to 1 line of configuration

