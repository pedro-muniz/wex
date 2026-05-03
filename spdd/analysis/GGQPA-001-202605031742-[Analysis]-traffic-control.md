# SPDD Analysis: Traffic Control Strategy

## Original Business Requirement
2. Traffic Control Strategy

To actually "control" the traffic rather than just passing it through, you need these three mechanisms:

    Load Leveling: Even if 10,000 users hit the API at once, the Workers only pull what they can handle (e.g., 50 messages at a time). The queue grows, but the database doesn't crash.

    Backpressure (TTL & Max Length): Configure your RabbitMQ queue with x-max-length. If the queue gets too full, you can reject new requests at the API level (returning a 429 Too Many Requests) to protect system stability.

    Fair Dispatch: Use channel.basicQos(prefetchCount: 1). This ensures RabbitMQ doesn't send a new message to a worker until that worker has acknowledged the previous one, preventing one "slow" worker from holding up a batch of tasks.  

## Domain Concept Identification

#### Existing Concepts (from codebase)
- **RabbitMQ Queue**: Queue declaration currently exists in consumers (e.g., `transaction_jobs`, `sync_jobs`) without arguments.
- **Worker / Consumer**: Go routines reading from `channel.Consume(..., autoAck=true, ...)` in services.
- **API Publisher**: Publishes jobs to the queue via `RabbitMQPublisher.PublishJob` and returns HTTP responses via API controllers.

#### New Concepts Required
- **Queue Arguments**: Additional configuration for queue length limits (`x-max-length` and `x-overflow`).
- **Manual Acknowledgement**: Replacing `autoAck=true` with explicit message acknowledgement (`d.Ack(false)`) on success or definitive failure.
- **Quality of Service (QoS)**: Configuring `prefetchCount` at the channel level before consumption.
- **Rate Limit Error Response**: Handling publisher rejections natively and mapping them to HTTP 429 Too Many Requests in the API tier.

#### Key Business Rules
- Workers must pull exactly what they can process concurrently (prefetch limit).
- A message must only be acknowledged after successful or definitively terminal processing.
- The system must actively refuse new work (429) rather than crashing or infinitely buffering when queues are overloaded.

## Strategic Approach

#### Solution Direction
- Implement **Fair Dispatch** by setting `channel.Qos(1, 0, false)` before consuming, and change `autoAck` to `false` in `channel.Consume`. Add manual `d.Ack(false)` inside the worker processing loop.
- Implement **Backpressure** by adding `amqp.Table{"x-max-length": <LIMIT>, "x-overflow": "reject-publish"}` to `QueueDeclare` across all relevant queues. 
- Enable Publisher Confirms (`channel.Confirm(false)`) in the `api_service` publisher. When `reject-publish` is triggered by a full queue, the publisher must detect the `basic.nack` or error. The API controllers will translate this error into a `429 Too Many Requests` HTTP response.

#### Key Design Decisions
- **Queue Limits configuration**: Where should the max length limit be configured? 
  - *Trade-offs*: Hardcoding vs environment variables. Environment variables allow easier production tuning without recompilation.
  - *Recommendation*: Use an environment variable with a sensible default (e.g., 10000).
- **Publisher Confirms vs basic.return**: To handle `reject-publish`, the publisher channel needs to wait for confirmation.
  - *Trade-offs*: Confirm mode adds slight latency but guarantees the API service knows whether the message was accepted or rejected due to queue limits.
  - *Recommendation*: Enable publisher confirms in `api_service` and implement a synchronous wait for the ack/nack to reliably detect when a queue is full.

#### Alternatives Considered
- *API-level token bucket rate limiting*: Rejected because it requires distributed state (like Valkey/Redis) and doesn't directly map to the actual queue depth backlog. Queue-based backpressure directly correlates with worker health and backlog depth.

## Risk & Gap Analysis

#### Requirement Ambiguities
- The requirement mentions TTL (Time to Live) alongside Max Length but doesn't specify a strategy (e.g., dead-lettering expired messages or just dropping them). We will focus on `x-max-length` for backpressure as requested.
- The specific `max-length` value is not prescribed in the requirement.

#### Edge Cases
- What happens if a worker crashes before acknowledging a message? RabbitMQ will automatically requeue it when the connection drops, which is correct, but processing must be idempotent.
- What if the queue declaration arguments change? RabbitMQ throws an error if trying to redeclare an existing durable queue with different arguments.

#### Technical Risks
- **Redeclaration Error**: Changing queue parameters (like `x-max-length`) on existing durable queues will cause `amqp091-go` to fail on startup. *Mitigation direction*: Document that existing queues must be deleted via RabbitMQ management interface during deployment, or dynamically alter them using RabbitMQ policies instead of hardcoded declaration arguments.
- **Publisher Throughput**: Enabling publisher confirms might reduce API throughput due to waiting for network round-trips. *Mitigation direction*: Monitor performance; if latency is too high, consider asynchronous confirm handling (though it complicates HTTP response mapping).

#### Acceptance Criteria Coverage
| AC# | Description | Addressable? | Gaps/Notes |
|-----|-------------|--------------|------------|
| 1 | Load leveling (queue acts as buffer) | Yes | Native to RabbitMQ |
| 2 | Configure queue with `x-max-length` and return 429 | Yes | Requires publisher confirms and API error mapping |
| 3 | Use `channel.basicQos(prefetchCount: 1)` and explicit ack | Yes | Requires modifying consumer configuration and logic |
