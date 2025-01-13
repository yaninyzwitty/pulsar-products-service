# PULSAR OUTBOX PRODUCTS SERVICE

This is a microservice aimed to show you how you design an outbox pattern to ensure idempotency, reliable message delivery by decoupling process of updating a database from sending events. This pattern often reliable for distributed systems therefore ensuring consistency.

<details>
<summary>Reasons for implementing the outbox pattern with pulsar<summary>

I wanted to refine my learning experience and improve my knowledge on designing realistic event driven architectures, especially with pulsar which is deemed to be better than kafka (it has a lower latency) in some cases. Also i wanted to be more informative on implementing transactions in golang, furthermore the outbox pattern separates database operations from event publishing, which ensures consistency.Moreover, I chose golang, instead of some other language, like typescript because I wanted to take advantage of its concurrency pattern, It worked really well.


## FEATURES

- Event-Driven Architecture: Leveraged Apache Pulsar for asynchronous message production, enabling a scalable and responsive system.
- Outbox Pattern: Decoupled database operations from event publishing to ensure consistency and seamless communication between services.
- Transactional Event Publishing: Integrated database operations with Pulsar to guarantee consistency and atomicity of data and event publishing.
- Concurrency: Utilized Golang's concurrency model for running the server and enabling non-blocking processing of outbox events efficiently.
- Idempotency: Implemented mechanisms to ensure messages (events) are processed only once, preventing duplication in distributed systems.



## TECH STACK

- **Language**: [Golang]
- **Server and router**: [Chi](https://github.com/go-chi/chi/)
- **Database**: [Neon Postgres](https://neon.tech/)
- **Postgres driver**: [Pgxpool](https://github.com/jackc/pgx)
- **ID generator**: [Snowflake](https://github.com/sony/sonyflake)
- **Message broker**: [Pulsar](https://www.datastax.com/products/astra-streaming/)


## Getting Started

```bash
git clone https://github.com/yaninyzwitty/pulsar-products-service.git
cd pulsar-products-service
go mod tidy
go run cmd/server/main.go
```

## Running Locally

Use the included `.env.local` to create `.env` file:

## Environmental Variables

Ensure that you configure all the necessary environmental variables and properly format the `config.yaml` file for the application to work correctly.

### Required Environmental Variables
- **`DB_PASSWORD`**: The password for your Neon database.  
- **`PULSAR_TOKEN`**: The token for your Astra Streaming (Pulsar) instance.

---

### Example `config.yaml` Format

```yaml
server:
  port: 3000  # The port the server listens on

database:
  username: neondb_owner  # The username for the Neon database
  host: ep-dark-pooler.eu-central-1.aws.neon.tech  # Neon database host
  database: neondb  # Name of the Neon database
  ssl_mode: require  # SSL mode for database connection

pulsar:
  uri: pulsar+ssl://pulsar-aws-eucentral1.streaming.datastax.com:6651  # Pulsar service URI
  topic_name: persistent://your-tenant/default/your_topic  # Name of the Pulsar topic
  token: some_token  # Pulsar authentication token


## Testing the Outbox Pattern

To test the implementation of the **Outbox Pattern**, you can use **Postman** or **cURL** to interact with the API. Below are the steps and details:

### Endpoint

### Request Body
Send a JSON payload representing the `Product` struct. Here’s the structure:

```json
{
  "name": "string",        // Name of the product
  "description": "string", // Detailed description of the product
  "price": 0.0,            // Price of the product
  "stock": 0,              // Number of items available in stock
  "category": "string",    // Category the product belongs to
  "created_at": "string",  // Timestamp for when the product was created (ISO format)
  "updated_at": "string"   // Timestamp for when the product was last updated (ISO format)
}

Using curl: 
curl -X POST http://localhost:3000/ \
-H "Content-Type: application/json" \
-d '{
  "name": "Sample Product",
  "description": "This is a test product",
  "price": 19.99,
  "stock": 100,
  "category": "Electronics",
  "created_at": "2025-01-13T10:00:00Z",
  "updated_at": "2025-01-13T10:00:00Z"
}'



