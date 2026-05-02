# SPDD Analysis: [GGQPA-XXX-202605021557] - 3-Service Microservices Architecture

## 1. Problem Statement
Decompose the Purchase Transaction system into three distinct services: API Consumer (Ingress), CRUD (Worker & Persistence), and Currency Conversion (Reporting). Each service must have clear boundaries and utilize shared infrastructure (RabbitMQ, Valkey, Postgres).

## 2. Requirements Analysis

### Microservice 1: API Consumer (Public Gateway)
- **POST /transactions**: 
    - Input: `description`, `transaction_date`, `purchase_amount`.
    - Action: Validate, Generate UUID, Store in Valkey with `PENDING` status, Publish to RabbitMQ.
- **GET /transactions/{id}/status**: 
    - Action: Retrieve status from Valkey.
    - Statuses: `PENDING`, `PROCESSING`, `COMPLETED`, `FAILED`.

### Microservice 2: Transaction CRUD (Persistence)
- **Consumer**: 
    - Action: Consume RabbitMQ message -> Validate -> Persist to Postgres -> Update Valkey status.
- **CRUD Endpoints**: 
    - `GET /transactions/{id}`, `PUT /transactions/{id}`, `DELETE /transactions/{id}`.
- **Infrastructure**: Injects `PostgresDAO` and `RabbitMQDAO`.

### Microservice 3: Currency Conversion (Reporting)
- **GET /transactions/{id}/convert?currency={currencyCode}**:
    - Integration: Treasury Reporting Rates of Exchange API.
    - Logic: Use rate <= purchase date within 6 months.
    - Output: Enriched DTO including `purchaseAmountUSD`, `exchangeRate`, and `convertedAmount`.

## 3. Domain Entities
- **PurchaseTransaction**: Extended to include `Status`.
- **Status Enum**: `PENDING`, `PROCESSING`, `COMPLETED`, `FAILED`.
- **ConversionRate**: As defined by Treasury API schema.

## 4. Technical Constraints
- **Infrastructure**: RabbitMQ, PostgreSQL, Valkey.
- **API Documentation**: OpenAPI/Swagger required.
- **Testing**: Unit and Integration tests.
- **Deployment**: Docker Compose for local environment.

## 5. Next Steps
- Update `domain.PurchaseTransaction` and status handling.
- Implement the Treasury API client.
- Set up the 3 main entry points and Docker Compose.
- Update Wire DI for 3 separate configurations.
