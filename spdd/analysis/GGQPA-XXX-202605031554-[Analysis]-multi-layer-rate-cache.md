# SPDD Analysis: Multi-Layer Rate Cache

## Original Business Requirement
Design a resilient worker responsible for retrieving currency exchange rates using a multi-layer cache strategy.
 The worker must prioritize Valkey and Postgres before calling any third-party API.
Context
The system processes requests containing:
currency_code (e.g., USD, BRL, EUR)
date (ISO format: YYYY-MM-DD)
Architecture includes:
Valkey (fast, in-memory cache)
Postgres (persistent storage of historical rates)
External third-party exchange rate API
Core Flow (Strict Order)
Generate Cache Key
Format: rate:{currency_code}:{date}
Check Valkey (L1 Cache)
If found:
Return cached value immediately
Mark as cache_hit_l1
Do NOT query Postgres or API
Check Postgres (L2 Cache / Source of Truth)
Query by currency_code and date
If found:
Store result in Valkey (warm cache)
Return result
Mark as cache_hit_l2
Call Third-Party API (Fallback)
If not found in Valkey or Postgres:
Fetch rate from external API
Handle timeouts, retries, and failures
Persist and Cache
On successful API response:
Store in Postgres (idempotent insert/update)
Store in Valkey with TTL
Return result
Mark as cache_miss
Advanced Requirements
1. Cache Strategy
Valkey TTL should be configurable (e.g., shorter for current day, longer for historical data)
Support stale-while-revalidate:
Optionally return stale Valkey data while refreshing in background
2. Concurrency Control
Prevent duplicate API calls for the same key:
Use distributed locks in Valkey (SETNX or similar)
Or request coalescing
3. Idempotency
Ensure multiple identical requests do not create duplicate DB records
Use (currency_code, date) as unique constraint in Postgres
4. Error Handling
If API fails:
Return stale data if available (Valkey or Postgres)
Otherwise return a controlled error
Log all failures and retries

## Domain Concept Identification

### Existing Concepts (from codebase)
- **Valkey Payload Store**: Existing Redis/Valkey infrastructure configured via `go-redis/v9` in `conversion_service/src/infra/repositories/valkey_payload_store.go`.
- **currency_conversion_rates Table**: Exists in Flyway migration `V2__create_currency_conversion_rate_table.sql` but lacks a unique constraint for idempotency.
- **TreasuryRateProvider**: The current component making direct HTTP calls to the Treasury API without caching.
- **TreasuryAPIDAO**: DAO that queries the Treasury API.

### New Concepts Required
- **RateCacheService/Worker**: A new orchestration layer that handles the L1 -> L2 -> API fallback logic.
- **RatePostgresRepository**: A new Postgres repository to interact with the `currency_conversion_rates` table to fulfill the L2 cache requirement.
- **DistributedLockProvider**: A concept to manage Valkey `SETNX` locks for concurrency control.
- **Stale-While-Revalidate Engine**: A background worker mechanism (e.g., goroutines) to fetch fresh rates asynchronously while returning stale ones.

### Key Business Rules
- **L1/L2 Strict Order**: Valkey must be checked first, then Postgres, before hitting the external API.
- **Idempotency**: Postgres inserts must use an `ON CONFLICT` clause based on `(target_currency, rate_date)` to prevent duplicates.
- **Concurrency Control**: Concurrent requests for the same `rate:currency:date` must not trigger multiple external API calls.
- **Stale-While-Revalidate**: Return stale Valkey data if API fails or if refreshing in the background.

## Strategic Approach

### Solution Direction
The caching logic should be implemented using a Decorator pattern or a new Orchestrator service around the existing `TreasuryRateProvider`. A new PostgreSQL repository will be introduced to fetch and store rates locally as an L2 cache. Concurrency control will be managed via Valkey distributed locks (`go-redis` SETNX) rather than in-memory request coalescing to ensure safety across multiple instances. 

### Key Design Decisions
- **Idempotency Constraint**: 
  - *Trade-offs*: Creating a new Flyway migration (V4) is required since V2 lacks the `UNIQUE(target_currency, rate_date)` constraint. 
  - *Recommendation*: Add migration `V4__add_unique_constraint_to_currency_rates.sql` to support Postgres `INSERT ... ON CONFLICT DO NOTHING/UPDATE`.
- **Concurrency Control (Distributed Locks vs Request Coalescing)**:
  - *Trade-offs*: Request coalescing (e.g. Go `singleflight`) is fast but only works for a single pod. Distributed locks work across the entire cluster but require Redis logic.
  - *Recommendation*: Use Valkey Distributed Locks (`SETNX` with a short expiration) since it is explicitly requested and handles multi-pod deployments properly.
- **Stale-while-revalidate**:
  - *Trade-offs*: Spawning unbounded goroutines can cause memory leaks.
  - *Recommendation*: Use controlled goroutines for asynchronous fetching. If Valkey has expired but stale data is available (requires storing stale data with a separate "stale" flag or using a longer absolute TTL but a shorter "refresh" TTL), the application will return it immediately and fire a background task.

### Alternatives Considered
- **In-Memory Cache (Go map or BigCache)**: Rejected because Valkey is explicitly requested as the L1 cache.
- **Overwriting existing TreasuryRateProvider directly**: We will create a `RateCacheProvider` that delegates to `TreasuryRateProvider` to separate caching concerns from API fetching concerns.

## Risk & Gap Analysis

### Requirement Ambiguities
- **Configurable TTLs**: The requirement mentions "shorter for current day, longer for historical data". It is unclear how exactly these durations should be configured (e.g. environment variables or code logic).
- **Stale-while-revalidate implementation**: Valkey natively supports TTL. Implementing stale-while-revalidate means we either keep keys around longer than their logical expiry, or we store an explicit "stale_at" timestamp inside the JSON value.

### Edge Cases
- **API Complete Outage**: If Valkey and Postgres are empty and the external API is down, what default behavior is expected? A controlled error is required.
- **Lock Contention**: If a node crashes while holding a Valkey lock, subsequent requests might be blocked until the lock TTL expires. Lock TTL must be short (e.g. 10s).
- **Date Matching**: The Treasury API often does not have rates for weekends. The DB unique constraint is on `date`. We must ensure the `date` queried corresponds exactly to the business logic date, which currently falls back to previous available dates.

### Technical Risks
- **Database Idempotency Issue**: Currently, `V2__create_currency_conversion_rate_table.sql` only has `CREATE INDEX idx_currency_conversion_rates_lookup`. We must introduce a `UNIQUE` constraint to fulfill the idempotency requirement safely.
- **Race conditions in Stale-While-Revalidate**: Need to ensure background refreshes don't overwhelm the API (they should also acquire the distributed lock).

### Acceptance Criteria Coverage
| AC# | Description | Addressable? | Gaps/Notes |
|-----|-------------|--------------|------------|
| 1 | Check Valkey (L1), return if found | Yes | None |
| 2 | Check Postgres (L2), store in L1, return | Yes | None |
| 3 | Call API fallback, persist L2, L1 | Yes | None |
| 4 | Valkey Configurable TTL | Yes | Will need to add config logic based on date |
| 5 | Stale-while-revalidate | Yes | Requires storing meta-data in Valkey value to know when it is stale vs expired |
| 6 | Distributed Lock (SETNX) | Yes | None |
| 7 | Idempotency (Postgres unique constraint) | Yes | Requires new Flyway migration |
| 8 | Return stale data on API fail | Yes | Requires falling back to L2/stale L1 |
