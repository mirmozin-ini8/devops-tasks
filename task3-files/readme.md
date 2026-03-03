# Microservices Project — Technical Documentation

## Overview

This document covers the implementation of Components 1 and 2 of the microservices assignment: the development of `book-service` and `user-service`, along with their respective database configurations. Both services are written in Go using the Gin web framework and are backed by isolated PostgreSQL instances.

---

## Component 1: Microservices Development

### Technology Stack

| Layer | Choice |
|---|---|
| Language | Go 1.23+ |
| Web Framework | Gin |
| Database Driver | `lib/pq` (PostgreSQL) |
| Authentication | `golang-jwt/jwt/v5` |
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
- `UpdateUser` — updates username and/or email; protected by the auth middleware.
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

### Dockerfile — Both Services

Both services use an identical multi-stage Docker build pattern:

**Stage 1 — Builder (`golang:1.26.0-alpine`)**
- Copies `go.mod` and `go.sum` first and runs `go mod download` to cache dependencies as a separate layer.
- Copies the rest of the source and compiles a statically linked binary with `CGO_ENABLED=0` targeting `linux/amd64`.

**Stage 2 — Runtime (`alpine:3.23.3`)**
- Copies only the compiled binary from the builder stage.
- Adds `ca-certificates` for HTTPS support.
- Creates a non-root `appuser` and runs the binary under that user, following the principle of least privilege.
- Exposes the appropriate port and sets the binary as the entrypoint.

This approach produces a small final image that contains no Go toolchain or source code.

---

## Component 2: Database Configuration

### Approach

Each microservice has its own isolated PostgreSQL instance. They do not share a database server, which is consistent with the microservices principle of data isolation — each service owns its data and is the only service that accesses it directly.

| Service | Database Name | Port (host) |
|---|---|---|
| book-service | `books_db` | 5432 |
| user-service | `users_db` | 5433 |
| order-service | `orders_db` | 5434 (planned) |

### Schema Management

Schema creation is handled in-process at service startup via the `createTable()` function in each service's `database/db.go`. The queries use `CREATE TABLE IF NOT EXISTS`, making the startup idempotent — the table is only created if it does not already exist. This is sufficient for the current stage and avoids the overhead of a migration framework.

### Connection Handling

The standard `database/sql` package is used with the `lib/pq` driver. `database/sql` provides built-in connection pooling by default. The pool is used as-is without explicit tuning (max open connections, max idle connections), which is acceptable for local and development environments. These settings can be configured later for production use.

### Local and Docker Compose Testing

Both services were verified locally and through Docker Compose before this documentation was written. In the Docker Compose setup:
- Each PostgreSQL instance runs as a separate container with a named volume for persistence.
- Services connect to their respective databases using the service name as the hostname (e.g., `books_db`, `users_db`), which is resolved by Docker's internal DNS.
- Environment variables are passed to each service container, overriding any `.env` defaults.

---

## Notes

- The `handler/book.go` file referenced in the user-service task description in the provided notes appears to be a copy-paste error from the book-service — the correct file for user-service is `handler/user.go`.
- The `middleware/auth.go` file was named `middleware/auth,go` (comma instead of period) in the provided notes; the correct filename used in implementation is `auth.go`.
- The `/metrics` endpoint in both services currently returns a placeholder string. This will be replaced with actual Prometheus metrics instrumentation in a later component.