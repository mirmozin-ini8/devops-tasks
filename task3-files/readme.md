# Microservices Book Ordering - Documentation

This document covers the implementation of the book ordering microservices application: the development of `book-service`, `user-service`, and `order-service`, along with their respective database configurations, Dockerfiles, and Docker Compose setup. All three services are written in Go using the Gin web framework and are backed by isolated PostgreSQL instances.

---

## Component 1: Microservices Development

### Technology Stack

| Layer | Choice |
|---|---|
| Language | Go 1.23+ |
| Web Framework | Gin |
| Database Driver | `lib/pq` (PostgreSQL) |
| Authentication | `golang-jwt/jwt/v5` |
| Password Hashing | `golang.org/x/crypto/bcrypt` |
| Environment Config | `joho/godotenv` |
| Containerization | Docker (multi-stage build) |

---

### Task 1.1 — book-service

#### Purpose

Manages the book catalog. Exposes a RESTful API for full CRUD operations on books, backed by a dedicated PostgreSQL database (`books_db`).

#### Directory Structure

```
book-service/
├── main.go
├── handler/
│   └── book.go
├── model/
│   └── book.go
├── repository/
│   └── book.go
├── database/
│   └── db.go
└── Dockerfile
```

#### Layers and Responsibilities

**`model/book.go`**
Defines the core data structures used across the service:
- `Book` — maps to the database row, serialized to JSON in API responses.
- `CreateBookRequest` — used for `POST /books`, with validation bindings (title and author required, price must be > 0, stock >= 0).
- `UpdateBookRequest` — used for `PUT /books/{id}`, all fields are optional pointers to support partial updates.

**`database/db.go`**
Opens a PostgreSQL connection using environment variables, verifies the connection with a ping, and runs a `CREATE TABLE IF NOT EXISTS` statement on startup to ensure the `books` table exists. The table schema is:

```sql
CREATE TABLE IF NOT EXISTS books (
    id         SERIAL PRIMARY KEY,
    title      TEXT NOT NULL,
    author     TEXT NOT NULL,
    price      NUMERIC(10, 2) NOT NULL,
    stock      INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

This approach handles schema bootstrapping without a separate migration tool, which is appropriate for this stage of the project.

**`repository/book.go`**
Contains all SQL logic, keeping database access isolated from HTTP handling:
- `GetAllBooks` — queries all rows ordered by ID, returns an empty slice (not nil) if no books exist.
- `GetBookByID` — queries by primary key, returns `nil, nil` if not found (distinguished from a DB error).
- `CreateBook` — inserts a row and uses `RETURNING` to get the created record back in a single query.
- `UpdateBook` — fetches the existing record first, applies only the fields that are non-nil in the request, then executes an `UPDATE`.
- `DeleteBook` — deletes by ID and checks `RowsAffected` to distinguish "not found" from a successful delete.

**`handler/book.go`**
The HTTP layer. Each handler parses the request, calls the appropriate repository function, and writes a JSON response with the correct status code. Error cases handled:
- Invalid path parameter (400)
- Record not found (404)
- Database or internal errors (500)

Also includes `HealthCheck`, which pings the database and returns service status.

**`main.go`**
Entry point. Loads `.env` if present (falls back gracefully if not), initializes the database connection, registers all routes under a `/books` group, and starts the Gin server on the port specified by `BOOKS_SERVER_PORT` (default: `8080`).

#### API Endpoints

| Method | Path | Description |
|---|---|---|
| GET | `/health` | Service and DB health check |
| GET | `/metrics` | Metrics placeholder |
| GET | `/books` | List all books |
| GET | `/books/:id` | Get a specific book |
| POST | `/books` | Create a new book |
| PUT | `/books/:id` | Update a book (partial update supported) |
| DELETE | `/books/:id` | Delete a book |

#### Environment Variables

```
BOOKS_DB_HOST=books_db 
BOOKS_DB_PORT=5432
BOOKS_DB_USER=postgres
BOOKS_DB_PASSWORD=password
BOOKS_DB_NAME=books_db
BOOKS_SERVER_PORT=8080
```

---

### Task 1.2 — user-service

#### Purpose

Handles user registration, login, and profile management. Issues JWT tokens on successful login. A protected endpoint (`PUT /users/:id`) requires a valid token via an auth middleware.

#### Directory Structure

```
user-service/
├── main.go
├── handler/
│   └── user.go
├── model/
│   └── user.go
├── repository/
│   └── user.go
├── database/
│   └── db.go
├── middleware/
│   └── auth.go
└── Dockerfile
```

#### Layers and Responsibilities

**`model/user.go`**
Defines:
- `User` — core struct; `PasswordHash` is tagged with `json:"-"` to ensure it is never serialized into API responses.
- `RegisterRequest` — validates username length (3–30 chars), email format, and password minimum length (6 chars).
- `LoginRequest` — username and password, both required.
- `UpdateUserRequest` — optional pointer fields for username and email.
- `LoginResponse` — wraps the JWT token and the user object returned on successful login.

**`database/db.go`**
Same pattern as book-service. Reads connection parameters from environment variables and creates the `users` table on startup:

```sql
CREATE TABLE IF NOT EXISTS users (
    id            SERIAL PRIMARY KEY,
    username      TEXT NOT NULL UNIQUE,
    email         TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

`UNIQUE` constraints on `username` and `email` enforce uniqueness at the database level, which is relied upon in the registration handler.

**`repository/user.go`**
- `GetUserByID` — returns the user without the password hash (not selected in query).
- `GetUserByUsername` — selects the password hash as well, used during login to compare credentials.
- `CreateUser` — inserts a new user with the pre-hashed password, returns the created record.
- `UpdateUser` — fetches the existing record, applies non-nil fields, runs an `UPDATE`.

**`middleware/auth.go`**
The `RequireAuth` middleware:
1. Reads the `Authorization` header.
2. Expects the format `Bearer <token>`.
3. Parses and validates the JWT using the `SECRET_KEY` environment variable and HMAC signing method verification.
4. On success, sets `user_id` in the Gin context for downstream handlers.
5. Aborts with 401 on any failure (missing header, wrong format, invalid or expired token).

**`handler/user.go`**
Handles the HTTP layer for all user-related operations:
- `Register` — binds the request, hashes the password using `bcrypt`, calls the repository to create the user.
- `Login` — looks up the user by username, compares the provided password against the stored bcrypt hash, and on success generates a signed JWT with the user's ID as a claim.
- `GetUser` — returns public user fields for a given ID.
- `UpdateUser` — updates username and/or email; protected by the auth middleware. The handler also verifies that the authenticated user's ID matches the requested `:id` parameter, preventing one user from modifying another user's profile.
- `HealthCheck` — pings the database and returns service status.

**`main.go`**
Loads environment, connects to the database, and sets up routes. The `PUT /users/:id` route is protected with `middleware.RequireAuth`. Server listens on `USERS_SERVER_PORT` (default: `8081`).

#### API Endpoints

| Method | Path | Auth Required | Description |
|---|---|---|---|
| GET | `/health` | No | Service and DB health check |
| GET | `/metrics` | No | Metrics placeholder |
| POST | `/users` | No | Register a new user |
| GET | `/users/:id` | No | Get user details |
| POST | `/login` | No | Login and receive JWT token |
| PUT | `/users/:id` | Yes (Bearer JWT) | Update user profile |

#### Environment Variables

```
USERS_DB_HOST=users_db
USERS_DB_PORT=5433
USERS_DB_USER=postgres
USERS_DB_PASSWORD=password
USERS_DB_NAME=users_db
USERS_SERVER_PORT=8081
```

---

### Task 1.3 — order-service

#### Purpose

Handles order creation and retrieval. Communicates with `book-service` to verify book existence and stock availability, and with `user-service` to verify that the requesting user exists. All order endpoints require JWT authentication.

#### Directory Structure

```
order-service/
├── main.go
├── handler/
│   └── order.go
├── model/
│   └── order.go
├── repository/
│   └── order.go
├── database/
│   └── db.go
├── middleware/
│   └── auth.go
├── client/
│   ├── book_client.go
│   └── user_client.go
└── Dockerfile
```

#### Layers and Responsibilities

**`model/order.go`**
Defines:
- `Order` — core struct mapping to the database row, including `user_id`, `book_id`, `quantity`, `total_price`, and `status`.
- `CreateOrderRequest` — requires `book_id` (> 0) and `quantity` (> 0).
- `BookResponse` — the subset of book fields consumed from the book-service response (`id`, `title`, `price`, `stock`).
- `UserResponse` — the subset of user fields consumed from the user-service response (`id`, `name`, `email`).

**`database/db.go`**
Same pattern as the other services. Creates the `orders` table on startup:

```sql
CREATE TABLE IF NOT EXISTS orders (
    id          SERIAL PRIMARY KEY,
    user_id     INTEGER NOT NULL,
    book_id     INTEGER NOT NULL,
    quantity    INTEGER NOT NULL CHECK (quantity > 0),
    total_price NUMERIC(10,2) NOT NULL,
    status      TEXT NOT NULL DEFAULT 'confirmed',
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

A `CHECK` constraint enforces that quantity is always greater than zero at the database level.

**`repository/order.go`**
- `CreateOrder` — inserts a new order and uses `RETURNING` to retrieve the full record in a single round trip.
- `GetOrderByID` — queries by primary key, returns `nil, nil` if no matching row exists.
- `GetOrdersByUserID` — queries all orders for a given user, ordered by `created_at` descending, returns an empty slice (not nil) if no orders exist.

**`client/book_client.go`**
Makes an HTTP GET request to `{BOOK_SERVICE_URL}/books/{id}`. Returns `nil, nil` if the book service responds with 404 (book not found), returns an error for any other non-200 status, and decodes the response body into a `BookResponse` struct on success.

**`client/user_client.go`**
Same pattern as the book client. Makes an HTTP GET request to `{USER_SERVICE_URL}/users/{id}`. Returns `nil, nil` on 404, an error on other non-200 responses, and decodes into a `UserResponse` on success.

**`handler/order.go`**
- `CreateOrder` — retrieves the authenticated `user_id` from the Gin context (set by the auth middleware), binds the request body, calls the user and book clients to validate both exist, checks that the requested quantity does not exceed available stock, computes `total_price = book.Price * quantity`, and persists the order.
- `GetOrderByID` — fetches a single order by ID from the repository.
- `GetOrdersByUserID` — fetches all orders for a user; verifies that the authenticated user's ID matches the requested `:userid` path parameter to prevent access to another user's orders.
- `HealthCheck` — pings the database and returns service status.

**`middleware/auth.go`**
Identical in logic to the user-service middleware. Parses and validates the JWT from the `Authorization: Bearer <token>` header using the shared `SECRET_KEY`, and sets `user_id` in the Gin context on success.

**`main.go`**
Loads environment, connects to the database, and registers all routes. All `/orders` routes are grouped under `middleware.RequireAuth`. Server listens on `ORDERS_SERVER_PORT` (default: `8082`).

#### API Endpoints

| Method | Path | Auth Required | Description |
|---|---|---|---|
| GET | `/health` | No | Service and DB health check |
| GET | `/metrics` | No | Metrics placeholder |
| POST | `/orders` | Yes (Bearer JWT) | Create a new order |
| GET | `/orders/:id` | Yes (Bearer JWT) | Get a specific order |
| GET | `/orders/user/:userid` | Yes (Bearer JWT) | Get all orders for a user |

#### Environment Variables

```
ORDERS_DB_HOST=orders_db
ORDERS_DB_PORT=5434
ORDERS_DB_USER=postgres
ORDERS_DB_PASSWORD=password
ORDERS_DB_NAME=orders_db
ORDERS_SERVER_PORT=8082
BOOK_SERVICE_URL=http://book-service:8080
USER_SERVICE_URL=http://user-service:8081
```

---

## Component 2: Database Configuration

### Approach

Each microservice has its own isolated PostgreSQL instance. They do not share a database server, which is consistent with the microservices principle of data isolation — each service owns its data and is the only service that accesses it directly.

| Service | Database Name | Port (host) |
|---|---|---|
| book-service | `books_db` | 5432 |
| user-service | `users_db` | 5433 |
| order-service | `orders_db` | 5434 |

### Schema Management

Schema creation is handled in-process at service startup via the `createTable()` function in each service's `database/db.go`. The queries use `CREATE TABLE IF NOT EXISTS`, making the startup idempotent — the table is only created if it does not already exist. This is sufficient for the current stage and avoids the overhead of a migration framework.

### Connection Handling

The standard `database/sql` package is used with the `lib/pq` driver. `database/sql` provides built-in connection pooling by default. The pool is used as-is without explicit tuning (max open connections, max idle connections), which is acceptable for local and development environments. These settings can be configured later for production use.

### Local and Docker Compose Testing

All three services were verified locally and through Docker Compose before this documentation was written. In the Docker Compose setup:
- Each PostgreSQL instance runs as a separate container with a named volume for persistence.
- Services connect to their respective databases using the service name as the hostname (e.g., `books_db`, `users_db`, `orders_db`), which is resolved by Docker's internal DNS.
- Environment variables are passed to each service container, overriding any `.env` defaults.

---

## Component 3: Containerization

### Task 3.1 — Dockerfiles

Each service uses an identical multi-stage Docker build pattern. The approach is consistent across all three services, with only the binary name, exposed port, and base image tag varying between them.

**Stage 1 — Builder (`golang:1.26.0-alpine`)**
- Copies `go.mod` and `go.sum` first and runs `go mod download` to cache the dependency layer independently from source changes.
- Copies the rest of the source and compiles a statically linked binary using `CGO_ENABLED=0` targeting `linux/amd64`.

**Stage 2 — Runtime (`alpine:3.23.3`)**
- Copies only the compiled binary from the builder stage, excluding the Go toolchain and all source code from the final image.
- Adds `ca-certificates` to support outbound HTTPS connections (required by order-service when calling other services, and generally recommended).
- Creates a non-root `appuser` and switches to that user before the entrypoint, following the principle of least privilege.
- Exposes the relevant port and sets the compiled binary as the container entrypoint.

The resulting images are small and contain no build tooling, source, or unnecessary dependencies.

#### book-service Dockerfile

```dockerfile
FROM golang:1.26.0-alpine AS builder

WORKDIR /app/

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o book-service .

FROM alpine:3.23.3

WORKDIR /root/
RUN apk --no-cache add ca-certificates
COPY --from=builder /app/book-service ./
RUN adduser -D appuser
USER appuser

EXPOSE 8080
CMD ["./book-service"]
```

---

### Task 3.2 — Docker Compose Setup

The `docker-compose.yml` file defines the complete local development environment. It brings up three PostgreSQL containers (one per service), three application containers, configures inter-service networking, and mounts named volumes for database persistence.

#### Services Defined

| Container | Image / Build | Port (host:container) | Depends On |
|---|---|---|---|
| `books_db` | `postgres:13` | `5432:5432` | — |
| `users_db` | `postgres:13` | `5433:5432` | — |
| `orders_db` | `postgres:13` | `5434:5432` | — |
| `book-service` | `./book-service` | `8080:8080` | `books_db` (healthy) |
| `user-service` | `./user-service` | `8081:8081` | `users_db` (healthy) |
| `order-service` | `./order-service` | `8082:8082` | `orders_db` (healthy), `book-service` (started), `user-service` (started) |

#### docker-compose.yml

```yaml
name: microservices-application

services:
  users_db:
    image: postgres:13
    environment:
      POSTGRES_USER: ${USERS_DB_USER}
      POSTGRES_PASSWORD: ${USERS_DB_PASSWORD}
      POSTGRES_DB: ${USERS_DB_NAME}
    ports:
      - "${USERS_DB_PORT}:5432"
    volumes:
      - users_db_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 5s
      timeout: 5s
      retries: 5

  books_db:
    image: postgres:13
    environment:
      POSTGRES_USER: ${BOOKS_DB_USER}
      POSTGRES_PASSWORD: ${BOOKS_DB_PASSWORD}
      POSTGRES_DB: ${BOOKS_DB_NAME}
    ports:
      - "${BOOKS_DB_PORT}:5432"
    volumes:
      - books_db_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 5s
      timeout: 5s
      retries: 5

  orders_db:
    image: postgres:13
    environment:
      POSTGRES_USER: ${ORDERS_DB_USER}
      POSTGRES_PASSWORD: ${ORDERS_DB_PASSWORD}
      POSTGRES_DB: ${ORDERS_DB_NAME}
    ports:
      - "${ORDERS_DB_PORT}:5432"
    volumes:
      - orders_db_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 5s
      timeout: 5s
      retries: 5

  user-service:
    build: ./user-service/
    container_name: user-service
    ports:
      - "8081:${USERS_SERVER_PORT}"
    environment:
      USERS_DB_HOST: ${USERS_DB_HOST}
      USERS_DB_PORT: 5432
      USERS_DB_USER: ${USERS_DB_USER}
      USERS_DB_PASSWORD: ${USERS_DB_PASSWORD}
      USERS_DB_NAME: ${USERS_DB_NAME}
      USERS_SERVER_PORT: ${USERS_SERVER_PORT}
      SECRET_KEY: ${SECRET_KEY}
    depends_on:
      users_db:
        condition: service_healthy

  book-service:
    build: ./book-service
    container_name: book-service
    ports:
      - "8080:${BOOKS_SERVER_PORT}"
    environment:
      BOOKS_DB_HOST: ${BOOKS_DB_HOST}
      BOOKS_DB_PORT: 5432
      BOOKS_DB_USER: ${BOOKS_DB_USER}
      BOOKS_DB_PASSWORD: ${BOOKS_DB_PASSWORD}
      BOOKS_DB_NAME: ${BOOKS_DB_NAME}
      BOOKS_SERVER_PORT: ${BOOKS_SERVER_PORT}
    depends_on:
      books_db:
        condition: service_healthy

  order-service:
    build: ./order-service/
    container_name: order-service
    ports:
      - "8082:${ORDERS_SERVER_PORT}"
    environment:
      ORDERS_DB_HOST: ${ORDERS_DB_HOST}
      ORDERS_DB_PORT: 5432
      ORDERS_DB_USER: ${ORDERS_DB_USER}
      ORDERS_DB_PASSWORD: ${ORDERS_DB_PASSWORD}
      ORDERS_DB_NAME: ${ORDERS_DB_NAME}
      ORDERS_SERVER_PORT: ${ORDERS_SERVER_PORT}
      BOOK_SERVICE_URL: http://book-service:8080
      USER_SERVICE_URL: http://user-service:8081
      SECRET_KEY: ${SECRET_KEY}
    depends_on:
      orders_db:
        condition: service_healthy
      book-service:
        condition: service_started
      user-service:
        condition: service_started

volumes:
  books_db_data:
  users_db_data:
  orders_db_data:

networks:
  default:
    driver: bridge
```

#### Running the Stack Locally

Build and start in detached mode:

```bash
docker compose up --build -d
```

Tear down containers and remove volumes:

```bash
docker compose down -v
```

Once running, the services are accessible at:
- book-service: `http://localhost:8080`
- user-service: `http://localhost:8081`
- order-service: `http://localhost:8082`