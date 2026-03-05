# Book Ordering System

A microservices-based book ordering application built with Go (Gin), deployed on Kubernetes using Helm. Three independent services communicate over HTTP using Kubernetes internal DNS, with all external traffic routed through an NGINX Ingress Controller.

---

## Table of Contents

- [Architecture Overview](#architecture-overview)
- [Services](#services)
- [Infrastructure](#infrastructure)
- [Microservices Development](#microservices-development)
- [Containerization](#containerization)
- [Kubernetes Deployment](#kubernetes-deployment)
- [Helm Chart](#helm-chart)
- [Deployment Procedure](#deployment-procedure)
- [API Testing](#api-testing)
- [Known Limitations](#known-limitations)
- [Debugging Reference](#debugging-reference)

---

## Architecture Overview

```
Client → book-ordering.dynv6.net
       → Jump Server (public IP)
         → kubectl port-forward :8080 → ingress-nginx-controller:80
           → Ingress (path-based routing)
               /books/*   → book-service:8080
               /users/*   → user-service:8081
               /login     → user-service:8081
               /orders/*  → order-service:8082
```

### Inter-Service Communication

`order-service` calls `book-service` and `user-service` over HTTP using Kubernetes internal DNS. URLs are injected via ConfigMap:

```
BOOK_SERVICE_URL = http://book-service:8080
USER_SERVICE_URL = http://user-service:8081
```

---

## Services

| Service | Port | Database | Image |
|---|---|---|---|
| book-service | 8080 | books_db | justnotmirr/book-service:v1.3 |
| user-service | 8081 | users_db | justnotmirr/user-service:v1.3 |
| order-service | 8082 | orders_db | justnotmirr/order-service:v1.3.1 |
| books-db | 5432 | books_db | postgres:15 |
| users-db | 5432 | users_db | postgres:15 |
| orders-db | 5432 | orders_db | postgres:15 |

---

## Infrastructure

| Component | Value |
|---|---|
| Master node | mozin-masternode |
| Worker node | mozin-workernode |
| Kubernetes | v1.31.14 (kubeadm) |
| CNI | Flannel |
| Ingress | NGINX (ingress-nginx) |
| Storage | rancher/local-path |
| Helm | v4.1.1 | 
| Domain | book-ordering.dynv6.net |

---

## Microservices Development

### Technology Stack

| Layer | Choice |
|---|---|
| Language | Go 1.22+ |
| Web Framework | Gin |
| Database driver | lib/pq (PostgreSQL) |
| Authentication | golang-jwt/jwt/v5 |
| Password hashing | golang.org/x/crypto/bcrypt |
| Env config | joho/godotenv |
| Containerization | Docker multi-stage builds |

---

### book-service

Manages the book catalog. Full CRUD on books backed by `books_db`.

**API Endpoints**

| Method | Path | Auth | Description |
|---|---|---|---|
| GET | /books/health | No | Health check |
| GET | /books/metrics | No | Metrics placeholder |
| GET | /books | No | List all books |
| GET | /books/:id | No | Get a specific book |
| POST | /books | No | Create a new book |
| PUT | /books/:id | No | Partial update |
| DELETE | /books/:id | No | Delete a book |

**Database Schema**

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

**Environment Variables**

```
BOOKS_DB_HOST=books-db
BOOKS_DB_PORT=5432
BOOKS_DB_USER=postgres
BOOKS_DB_PASSWORD=<from secret>
BOOKS_DB_NAME=books_db
BOOKS_SERVER_PORT=8080
```

---

### user-service

Handles user registration, login, and profile management. Issues JWT tokens on login.

**API Endpoints**

| Method | Path | Auth | Description |
|---|---|---|---|
| GET | /users/health | No | Health check |
| GET | /users/metrics | No | Metrics placeholder |
| POST | /users | No | Register a new user |
| GET | /users/:id | No | Get user details |
| POST | /login | No | Login — returns JWT token |
| PUT | /users/:id | Yes — Bearer JWT | Update own profile only |

**Database Schema**

```sql
CREATE TABLE IF NOT EXISTS users (
    id            SERIAL PRIMARY KEY,
    username      TEXT NOT NULL UNIQUE,
    email         TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**JWT Authentication Flow**

1. Client POSTs credentials to `POST /login`
2. user-service verifies password against bcrypt hash
3. On success, signs a JWT containing `user_id` using `SECRET_KEY`
4. Client includes token in subsequent requests: `Authorization: Bearer <token>`
5. Auth middleware validates signature, extracts `user_id`, sets in Gin context
6. Handlers read `user_id` from context — never from the request body

> **Security note:** `user_id` is always taken from the validated JWT, never from the request body. This prevents users from creating or accessing resources belonging to other users.

**Environment Variables**

```
USERS_DB_HOST=users-db
USERS_DB_PORT=5432
USERS_DB_USER=postgres
USERS_DB_PASSWORD=<from secret>
USERS_DB_NAME=users_db
USERS_SERVER_PORT=8081
SECRET_KEY=<from secret>
```

---

### order-service

Handles order creation and retrieval. Calls `book-service` to verify availability and `user-service` to verify the user exists. All order endpoints require JWT authentication.

**API Endpoints**

| Method | Path | Auth | Description |
|---|---|---|---|
| GET | /orders/health | No | Health check |
| GET | /orders/metrics | No | Metrics placeholder |
| POST | /orders | Yes — Bearer JWT | Create a new order |
| GET | /orders/user/:userId | Yes — Bearer JWT | Get all orders for authenticated user |
| GET | /orders/:id | Yes — Bearer JWT | Get a specific order |

**Order Creation Flow**

1. Client sends `POST /orders` with `book_id` and `quantity` (JWT required)
2. Middleware validates JWT, extracts `user_id`
3. Calls `GET /users/:userId` to verify user exists
4. Calls `GET /books/:bookId` to verify book and check stock
5. If `book.Stock < quantity` → returns `400` with available stock
6. Computes `total_price = book.Price * quantity`
7. Inserts order with status `confirmed`

**Database Schema**

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

**Route Order (important)**

Static route MUST come before dynamic route in Gin
```go
protected.GET("/user/:userId", handler.GetOrdersByUserID) 
protected.GET("/:id", handler.GetOrderByID)   
```

**Environment Variables**

```
ORDERS_DB_HOST=orders-db
ORDERS_DB_PORT=5432
ORDERS_DB_USER=postgres
ORDERS_DB_PASSWORD=<from secret>
ORDERS_DB_NAME=orders_db
ORDERS_SERVER_PORT=8082
BOOK_SERVICE_URL=http://book-service:8080
USER_SERVICE_URL=http://user-service:8081
SECRET_KEY=<from secret>
```

---

## Containerization

### Multi-Stage Dockerfile Pattern

All three services use the same multi-stage build - Stage 1 compiles, Stage 2 is a minimal Alpine runtime with only the binary. Example `Dockerfile` for  book-service:

```dockerfile
FROM golang:1.22-alpine AS builder

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

> `CGO_ENABLED=0` produces a fully static binary with no libc dependency, allowing it to run on minimal Alpine.

### Docker Compose (Local Development Only)

| Container | Host Port | Depends On |
|---|---|---|
| books_db | 5432:5432 | — |
| users_db | 5433:5432 | — |
| orders_db | 5434:5432 | — |
| book-service | 8080:8080 | books_db (healthy) |
| user-service | 8081:8081 | users_db (healthy) |
| order-service | 8082:8082 | orders_db (healthy), book-service, user-service |

> **Note:** Docker Compose uses different host ports (5432, 5433, 5434) to avoid conflicts on the local machine. In Kubernetes all three databases use port 5432 — they are isolated by separate pods and ClusterIP services.

```bash
# Start everything
docker compose up --build -d

# Tear down and remove volumes
docker compose down -v
```

---

## Kubernetes Deployment

### Storage Provisioner

If the cluster has no default `StorageClass`, it will cause all PVCs to remain in `Pending` state. Install `rancher/local-path-provisioner` to resolve this.

```bash
kubectl apply -f https://raw.githubusercontent.com/rancher/local-path-provisioner/master/deploy/local-path-storage.yaml
```

Set as default:

```bash
kubectl patch storageclass local-path \
  -p '{"metadata": {"annotations":{"storageclass.kubernetes.io/is-default-class":"true"}}}'
```

Verify:

```bash
kubectl get storageclass
```

> Without a StorageClass, PVCs stay `Pending` → databases never start → services crash-loop.

### Pod Startup Behaviour

Application services experience 1–2 restarts on first deploy. This is expected:

```
t=0s    All 6 pods start simultaneously
t=2s    Services try to connect to database — connection refused (postgres still initialising)
t=2s    Services crash (Exit Code 1)
t=30s   Databases ready
t=31s   Kubernetes restarts services — connect successfully
t=40s   All 6 pods Running
```

Kubernetes has no `depends_on`. Services are designed to crash-and-restart until dependencies are ready.

---

## Helm Chart

### Structure

```
book-ordering/
├── Chart.yaml
├── values.yaml
└── templates/
    ├── namespace.yaml
    ├── secrets.yaml
    ├── configmap.yaml
    ├── books-db.yaml
    ├── users-db.yaml
    ├── orders-db.yaml   
    ├── book-service.yaml  
    ├── user-service.yaml 
    ├── order-service.yaml 
    └── ingress.yaml
```

### Key `values.yaml` Settings

```yaml
images:
  bookService:
    repository: justnotmirr/book-service
    tag: v1.3
    pullPolicy: Always
  userService:
    repository: justnotmirr/user-service
    tag: v1.3
    pullPolicy: Always
  orderService:
    repository: justnotmirr/order-service
    tag: v1.3
    pullPolicy: Always
  postgres:
    repository: postgres
    tag: "15"
    pullPolicy: IfNotPresent

database:
  books:
    host: books-db 
    name: books_db
    user: postgres
    port: "5432" 

ingress:
  host: book-ordering.dynv6.net
  className: nginx
```

### Secrets

db-secret contains `base64` of postgres-password and the jwt-secret.

```yaml
postgres-password: cGFzc3dvcmQ= 
jwt-secret: aTV1MGd5VVJKcg==  
```

To generate a base64 value:

```bash
echo -n 'your-value' | base64
```

Secrets are injected per-service:

| Service | Env Var | Secret Key |
|---|---|---|
| book-service | BOOKS_DB_PASSWORD | postgres-password |
| user-service | USERS_DB_PASSWORD | postgres-password |
| user-service | SECRET_KEY | jwt-secret |
| order-service | ORDERS_DB_PASSWORD | postgres-password |
| order-service | SECRET_KEY | jwt-secret |
| All databases | POSTGRES_PASSWORD | postgres-password |

### Health Probes

| Service | Liveness Path | Readiness Path | Notes |
|---|---|---|---|
| book-service | /books/health | /books/health | HTTP GET on port 8080 |
| user-service | /users/health | /users/health | HTTP GET on port 8081 |
| order-service | /orders/health | /orders/health | HTTP GET on port 8082|
| *-db | pg_isready -U postgres | pg_isready -U postgres | exec probe |

> Health and metrics routes in `order-service` are registered **outside** the auth middleware group so Kubernetes probes don't require a JWT token.

### Ingress Routing

Path-based routing with regex:

```yaml
/books(/|$)(.*)   → book-service:8080    (ImplementationSpecific)
/users(/|$)(.*)   → user-service:8081    (ImplementationSpecific)
/login            → user-service:8081    (Exact)
/orders(/|$)(.*)  → order-service:8082   (ImplementationSpecific)
```

---

## Deployment Procedure

### 1. Build and Push Images

Build (local machine):
```powershell
docker build -t justnotmirr/book-service:v1.3 ./book-service
docker build -t justnotmirr/user-service:v1.3 ./user-service
docker build -t justnotmirr/order-service:v1.3 ./order-service
```

Push to Docker Hub:

```powershell
docker push justnotmirr/book-service:v1.3
docker push justnotmirr/user-service:v1.3
docker push justnotmirr/order-service:v1.3
```

### 2. Copy Chart to Jump Server

```powershell
scp -r .\book-ordering\ azureuser@<jump-server-public-ip>:~/
```

### 3. Validate

```bash
helm lint ./book-ordering/
helm install book-ordering ./book-ordering/ --dry-run=client --debug
```

### 4. Deploy

```bash
helm install book-ordering ./book-ordering/

kubectl get pods -n book-ordering -w
```

### 5. Upgrade After Changes

Update the image tag in values.yaml, copy chart, then run:

```bash
helm upgrade book-ordering ./book-ordering/
```

Verify the right image is running:

```bash
kubectl get deployment book-service -n book-ordering \
  -o jsonpath='{.spec.template.spec.containers[0].image}'
```

### 6. Access

Port-forward ingress on jump server:

```bash
kubectl port-forward -n ingress-nginx svc/ingress-nginx-controller 8080:80 &
```
Test with Host header or directly via domain:

```bash
curl -H "Host: book-ordering.dynv6.net" http://localhost:8080/books/health

curl http://book-ordering.dynv6.net/books/health
```

### 7. To Uninstall

```bash
helm uninstall book-ordering
kubectl delete pvc -n book-ordering --all
kubectl delete namespace book-ordering
```

---

## API Testing

**Step 1 — Health Checks**

```
GET http://book-ordering.dynv6.net/books/health
GET http://book-ordering.dynv6.net/users/health
GET http://book-ordering.dynv6.net/orders/health
```

**Step 2 — Register a User**

```
POST http://book-ordering.dynv6.net/users
Content-Type: application/json

{
  "username": "<username>",
  "email": "<email>",
  "password": "<password>"
}
```

**Step 3 — Login and Copy Token**

Login using the created username and password. Copy the token from the response.
```
POST http://book-ordering.dynv6.net/login
Content-Type: application/json

{
  "username": "<username>",
  "password": "<password>"
}
```
Response: `{ "token": "eyJhbGc...", "user": { "id": 1, ... } }`

**Step 4 — Create a Book**

```
POST http://book-ordering.dynv6.net/books
Content-Type: application/json

{
  "title": "<book-name>",
  "author": "<Alan Donovan>",
  "price": <49.99>,
  "stock": <100>
}
```

**Step 5 — Create an Order**

```
POST http://book-ordering.dynv6.net/orders
Authorization: Bearer eyJhbGc...
Content-Type: application/json

{
  "book_id": <1>,
  "quantity": <2>
}
```

**Step 6 — Get Orders for User**

```
GET http://book-ordering.dynv6.net/orders/user/1
Authorization: Bearer eyJhbGc...
```

1 = your user_id from the login response.
Using another user's ID returns 403 Forbidden.

### Common Error Responses

| Code | Error | Cause | Fix |
|---|---|---|---|
| 401 | authorization header required | Missing Bearer token | Add `Authorization: Bearer <token>` header |
| 403 | cannot view another user's orders | JWT user_id != URL user ID | Use your own user ID |
| 400 | insufficient stock | Quantity > book.Stock | Reduce quantity or increase stock |
| 503 | user/book service unavailable | order-service can't reach dependency | Check pod logs and DNS |

---

## Known Limitations

- **Init containers** — Services crash 1-2 times on first deploy waiting for postgres.
- **Database migrations** — Schema is created via `CREATE TABLE IF NOT EXISTS` on startup.
- **Secret management** — Secrets are base64-encoded in the Helm chart.

---

## Debugging Reference

Pod status:

```bash
kubectl get pods -n book-ordering
kubectl get pods -n book-ordering -w
```

Logs:

```bash
kubectl logs -n book-ordering deployment/book-service
kubectl logs -n book-ordering deployment/user-service
kubectl logs -n book-ordering deployment/order-service
```

Verify routes actually registered:

```bash
kubectl logs -n book-ordering deployment/order-service | grep "GIN-debug"
```

Verify image tag deployed:

```bash
kubectl get deployment book-service -n book-ordering \
  -o jsonpath='{.spec.template.spec.containers[0].image}'
```

Pod details and events:

```bash
kubectl describe pod <pod-name> -n book-ordering
```

PVC and storage:

```bash
kubectl get pvc -n book-ordering
kubectl get storageclass
```

Services and ingress:

```bash
kubectl get svc -n book-ordering
kubectl get ingress -n book-ordering
```

Helm:

```bash
helm list
helm history book-ordering
helm rollback book-ordering 1
```