# SPDD Analysis: E2E Tests Implementation for Store and Convert Transactions

## Original Business Requirement
Write E2E tests in ./e2e_tests/ folderfor the following scenarios:

Feature: Store Purchase Transaction
  Scenario: Successfully store a valid transaction
    Given a transaction with description "Book purchase"
    And a valid transaction date "2026-05-01"
    And a purchase amount of 25.50 USD
    When the transaction is submitted
    Then the transaction should be stored
    And a unique identifier should be generated

  Scenario: Reject transaction with description longer than 50 characters
    Given a transaction with description exceeding 50 characters
    When the transaction is submitted
    Then the system should return a validation error

  Scenario: Reject transaction with invalid date
    Given a transaction with an invalid date format
    When the transaction is submitted
    Then the system should return a validation error

  Scenario: Reject transaction with negative amount
    Given a transaction with a negative purchase amount
    When the transaction is submitted
    Then the system should return a validation error

Feature: Store Purchase Transaction - Edge Cases
  Scenario: Accept description with exactly 50 characters
    Given a transaction with a description of exactly 50 characters
    And a valid transaction date "2026-05-01"
    And a purchase amount of 10.00 USD
    When the transaction is submitted
    Then the transaction should be stored successfully

  Scenario: Reject description with 51 characters
    Given a transaction with a description of 51 characters
    When the transaction is submitted
    Then the system should return a validation error

  Scenario: Reject transaction with zero amount
    Given a transaction with a purchase amount of 0.00 USD
    When the transaction is submitted
    Then the system should return a validation error

  Scenario: Accept transaction with minimum valid positive amount
    Given a transaction with a purchase amount of 0.01 USD
    And a valid transaction date "2026-05-01"
    When the transaction is submitted
    Then the transaction should be stored successfully

  Scenario: Round purchase amount to nearest cent
    Given a transaction with a purchase amount of 10.005 USD
    When the transaction is submitted
    Then the stored amount should be 10.01 USD
    
Feature: Retrieve Transaction with Currency Conversion
  Scenario: Successfully retrieve transaction in another currency
    Given a stored transaction with ID "123"
    And a valid exchange rate available for the transaction date
    When the transaction is requested in "EUR"
    Then the system should return the transaction details
    And include the original USD amount
    And include the exchange rate used
    And include the converted amount in EUR

  Scenario: Use closest exchange rate within 6 months
    Given a stored transaction with ID "123"
    And no exact exchange rate for the transaction date
    But a valid rate exists within the last 6 months
    When the transaction is requested in "EUR"
    Then the system should use the closest previous rate

  Scenario: Fail when no exchange rate is available within 6 months
    Given a stored transaction with ID "123"
    And no exchange rate within 6 months before the transaction date
    When the transaction is requested in "EUR"
    Then the system should return an error indicating conversion is not possible

  Scenario: Converted amount should be rounded to two decimals
    Given a stored transaction with ID "123"
    And a valid exchange rate
    When the transaction is requested in "EUR"
    Then the converted amount should be rounded to two decimal places

Feature: Retrieve Transaction - Edge Cases
  Scenario: Use exchange rate exactly on 6-month boundary
    Given a stored transaction with ID "123"
    And an exchange rate exactly 6 months before the transaction date
    When the transaction is requested in "EUR"
    Then the system should use this exchange rate

  Scenario: Reject conversion if rate is older than 6 months
    Given a stored transaction with ID "123"
    And the only available exchange rate is older than 6 months
    When the transaction is requested in "EUR"
    Then the system should return a conversion error

  Scenario: Handle very large purchase amounts
    Given a stored transaction with a very large amount (e.g., 1000000000.00 USD)
    And a valid exchange rate
    When the transaction is requested in "EUR"
    Then the conversion should be calculated correctly without overflow

  Scenario: Handle rounding of converted amount
    Given a stored transaction with ID "123"
    And an exchange rate that results in more than two decimal places
    When the transaction is requested in "EUR"
    Then the converted amount should be rounded to two decimal places

  Scenario: Reject request for unsupported currency
    Given a stored transaction with ID "123"
    When the transaction is requested in an unsupported currency
    Then the system should return an error

## Domain Concept Identification

#### Existing Concepts (from codebase)
- **PurchaseTransaction**: Main entity for storing transaction records. Fields: ID, Description, Amount, Date, Status.
- **CurrencyConversion**: Process of converting USD amounts to target currencies using historical rates.
- **ExchangeRate**: Data from Treasury API, cached in Postgres and Valkey.

#### New Concepts Required
- **E2E Test Suite**: A new conceptual layer in the project to orchestrate full-flow testing across multiple microservices.
- **Environment Orchestration**: Concepts for managing service state (DB seeding, Valkey clearing) during test execution.

#### Key Business Rules
- **Validation Rules**: Description <= 50, Amount > 0.
- **Rounding Rules**: Nearest cent (2 decimals) for both USD and converted values.
- **Conversion Rule**: 6-month lookback for exchange rates.
- **E2E Goal**: All microservices (API, Transaction, Conversion) must interact correctly to fulfill the requirements.

## Strategic Approach

#### Solution Direction
- **Tooling**: Use standard Go testing tools (`go test`) combined with HTTP clients to simulate external requests.
- **Folder Structure**: Create a dedicated `./e2e_tests/` directory at the project root.
- **Testing Strategy**:
    - **Black-box Testing**: Interact with the system via the `api_service` endpoints.
    - **State Verification**: Directly check the DB or Valkey if necessary, or use the API's status/retrieval endpoints to verify outcomes.
    - **Mocking/Seeding**: Use the existing Flyway migrations and potentially a test-specific seeding mechanism for exchange rates to test the 6-month rule.

#### Key Design Decisions
- **Stand-alone Test Suite**: The E2E tests will reside in their own package/folder to avoid circular dependencies and clearly separate unit tests from system tests.
- **Container-Aware**: Tests should be able to run against the docker-compose environment or a locally running stack.
- **Data Isolation**: Each test scenario should ideally use unique data (UUIDs, unique descriptions) to avoid interference.

#### Alternatives Considered
- **Integration tests within each service**: Rejected because the requirement explicitly asks for a dedicated `./e2e_tests/` folder for cross-service scenarios.

## Risk & Gap Analysis

#### Requirement Ambiguities
- **Test Framework**: No specific E2E/BDD framework was requested (e.g., Gherkin/Godog). Defaulting to standard Go tests for maximum compatibility unless specified.
- **Service Orchestration**: Should the tests start/stop services, or assume they are already running? Recommendation: Assume a running environment (e.g., via `docker-compose up`).

#### Edge Cases
- **Async Latency**: E2E tests must account for the asynchronous nature of the Transaction and Conversion services (using retries or polling).
- **Time Sensitivity**: Testing the 6-month rule requires careful date manipulation or controlled seeding of the `currency_conversion_rates` table.

#### Technical Risks
- **Flakiness**: Asynchronous event-driven systems are prone to flaky tests due to timing issues. Mitigation: Robust polling with timeouts.
- **Database Pollution**: Successive test runs might fail if the database is not cleaned or if constraints are violated.

#### Acceptance Criteria Coverage
| AC# | Description | Addressable? | Gaps/Notes |
|-----|-------------|--------------|------------|
| 1 | Store valid transaction | Yes | POST /transactions and verify result. |
| 2 | Validation errors (len, date, neg) | Yes | Verify 400 Bad Request responses. |
| 3 | Description boundary (50/51) | Yes | Verify edge case validation. |
| 4 | Amount boundary (0/0.01) | Yes | Verify edge case validation. |
| 5 | USD Rounding | Yes | Verify stored/returned amount is rounded. |
| 6 | Conversion retrieval | Yes | POST /convert and GET /convert result. |
| 7 | 6-month rate selection | Yes | Requires seeding rates at specific dates. |
| 8 | 6-month boundary match | Yes | Requires precise seeding. |
| 9 | Conversion failure (> 6 months) | Yes | Verify error response when rates are old. |
| 10 | Converted Rounding | Yes | Verify converted amount has 2 decimals. |
| 11 | Unsupported currency | Yes | Verify error response. |
