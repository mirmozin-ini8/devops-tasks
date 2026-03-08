# Book Ordering System — Complete Helm Chart
# Every file, every line explained

---

## Chart.yaml

This is the metadata file for your Helm chart. Every chart must have one.

```yaml
# helm/book-ordering/Chart.yaml

apiVersion: v2           # Helm chart API version — v2 is for Helm 3
name: book-ordering      # name of the chart — used in helm install/upgrade
description: Book ordering microservices system
type: application        # "application" (deployable) vs "library" (reusable templates)
version: 0.1.0           # chart version — increment when you change the chart itself
appVersion: "v1"         # your application version — matches your Docker image tags
```

---

## values.yaml

This is the most important file. All configurable values live here.
Templates reference these values using {{ .Values.xxx }} syntax.
To deploy a new version, you only change this file — never the templates.

```yaml
# helm/book-ordering/values.yaml

# ── Namespace ─────────────────────────────────────────────────────────
namespace: book-ordering

# ── Image settings ────────────────────────────────────────────────────
images:
  bookService:
    repository: justnotmirr/book-service
    tag: v1
    pullPolicy: Always     # Always pull — ensures you always get latest push
  userService:
    repository: justnotmirr/user-service
    tag: v1
    pullPolicy: Always
  orderService:
    repository: justnotmirr/order-service
    tag: v1
    pullPolicy: Always
  postgres:
    repository: postgres
    tag: "15"
    pullPolicy: IfNotPresent   # stable image — only pull if not cached

# ── Service ports ─────────────────────────────────────────────────────
ports:
  bookService: 8080
  userService: 8081
  orderService: 8082
  postgres: 5432

# ── Database configuration (non-sensitive) ────────────────────────────
# Passwords are in secrets.yaml, not here
database:
  books:
    name: books_db
    user: postgres
    host: books-db        # matches the Kubernetes Service name
    port: "5432"
  users:
    name: users_db
    user: postgres
    host: users-db
    port: "5432"
  orders:
    name: orders_db
    user: postgres
    host: orders-db
    port: "5432"

# ── Inter-service URLs ────────────────────────────────────────────────
# Inside Kubernetes, services reach each other by service-name.namespace
# Format: http://<service-name>.<namespace>.svc.cluster.local:<port>
# Shorthand (same namespace): http://<service-name>:<port>
serviceUrls:
  bookService: "http://book-service:8080"
  userService: "http://user-service:8081"

# ── Resource limits ───────────────────────────────────────────────────
# requests = guaranteed minimum Kubernetes reserves for this pod
# limits   = maximum the pod can use before being throttled/killed
# CPU unit: 1000m = 1 CPU core. 100m = 0.1 core
# Memory unit: Mi = mebibytes
resources:
  bookService:
    requests:
      cpu: "100m"
      memory: "128Mi"
    limits:
      cpu: "200m"
      memory: "256Mi"
  userService:
    requests:
      cpu: "100m"
      memory: "128Mi"
    limits:
      cpu: "200m"
      memory: "256Mi"
  orderService:
    requests:
      cpu: "100m"
      memory: "128Mi"
    limits:
      cpu: "200m"
      memory: "256Mi"
  postgres:
    requests:
      cpu: "100m"
      memory: "256Mi"
    limits:
      cpu: "300m"
      memory: "512Mi"

# ── Replica counts ────────────────────────────────────────────────────
replicaCount:
  bookService: 1
  userService: 1
  orderService: 1

# ── Ingress ───────────────────────────────────────────────────────────
ingress:
  host: book-ordering.ddns.net    # change this to your actual domain/DDNS
  className: nginx

# ── Persistent storage ────────────────────────────────────────────────
storage:
  books: "1Gi"
  users: "1Gi"
  orders: "1Gi"
```

---

## templates/namespace.yaml

Creates an isolated namespace for all your resources.
Everything related to this app lives in `book-ordering` namespace.

```yaml
# helm/book-ordering/templates/namespace.yaml

apiVersion: v1
kind: Namespace
metadata:
  name: {{ .Values.namespace }}
  labels:
    app: book-ordering
```

**`{{ .Values.namespace }}`** — Helm template syntax. At deploy time, Helm
replaces this with the value from values.yaml — "book-ordering".
The double curly braces are Go's template syntax, which Helm uses.

---

## templates/secrets.yaml

Stores sensitive values. Base64 encoded but NOT encrypted by default.
In production you'd use Sealed Secrets or Vault — for this assignment
base64 is fine.

```yaml
# helm/book-ordering/templates/secrets.yaml

apiVersion: v1
kind: Secret
metadata:
  name: db-secrets
  namespace: {{ .Values.namespace }}
type: Opaque      # Opaque = generic secret (vs kubernetes.io/tls for certs)
data:
  # Base64 encoded values
  # To encode: echo -n "password" | base64
  # "password" → "cGFzc3dvcmQ="
  # "HEYBROTHER" → "SEVZQlJPVEhFUg=="
  postgres-password: cGFzc3dvcmQ=    # "password"
  jwt-secret: SEVZQ1JPVEhFUg==       # "HEYBROTHER"
```

**How to generate your own base64 values** (run on jump server):
```bash
echo -n "password" | base64      # cGFzc3dvcmQ=
echo -n "HEYBROTHER" | base64    # SEVZQ1JPVEhFUg==
```

The `-n` flag is critical — without it echo adds a newline character
which gets encoded too, giving you a wrong value.

---

## templates/configmap.yaml

All non-sensitive configuration in one place.
Both the database services and application services reference this.

```yaml
# helm/book-ordering/templates/configmap.yaml

apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
  namespace: {{ .Values.namespace }}
data:
  # Book service database config
  BOOKS_DB_HOST: {{ .Values.database.books.host }}
  BOOKS_DB_PORT: {{ .Values.database.books.port | quote }}
  BOOKS_DB_USER: {{ .Values.database.books.user }}
  BOOKS_DB_NAME: {{ .Values.database.books.name }}
  BOOKS_SERVER_PORT: {{ .Values.ports.bookService | quote }}

  # User service database config
  USERS_DB_HOST: {{ .Values.database.users.host }}
  USERS_DB_PORT: {{ .Values.database.users.port | quote }}
  USERS_DB_USER: {{ .Values.database.users.user }}
  USERS_DB_NAME: {{ .Values.database.users.name }}
  USERS_SERVER_PORT: {{ .Values.ports.userService | quote }}

  # Order service database config
  ORDERS_DB_HOST: {{ .Values.database.orders.host }}
  ORDERS_DB_PORT: {{ .Values.database.orders.port | quote }}
  ORDERS_DB_USER: {{ .Values.database.orders.user }}
  ORDERS_DB_NAME: {{ .Values.database.orders.name }}
  ORDERS_SERVER_PORT: {{ .Values.ports.orderService | quote }}

  # Inter-service URLs
  BOOK_SERVICE_URL: {{ .Values.serviceUrls.bookService }}
  USER_SERVICE_URL: {{ .Values.serviceUrls.userService }}
```

**`| quote`** — a Helm template filter. Wraps the value in quotes.
Needed for numbers — YAML would treat `5432` as an integer but
your Go code reads it as a string with `os.Getenv()`.
`{{ .Values.ports.bookService | quote }}` produces `"8080"` not `8080`.

---

## templates/books-db.yaml

Deployment + Service for books PostgreSQL database.
PersistentVolumeClaim ensures data survives pod restarts.

```yaml
# helm/book-ordering/templates/books-db.yaml

# ── PersistentVolumeClaim ──────────────────────────────────────────────
# Requests storage from the cluster
# Data persists even if the pod is deleted and recreated
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: books-db-pvc
  namespace: {{ .Values.namespace }}
spec:
  accessModes:
    - ReadWriteOnce     # can be mounted by one node at a time (fine for postgres)
  resources:
    requests:
      storage: {{ .Values.storage.books }}

---
# ── Deployment ────────────────────────────────────────────────────────
apiVersion: apps/v1
kind: Deployment
metadata:
  name: books-db
  namespace: {{ .Values.namespace }}
  labels:
    app: books-db
spec:
  replicas: 1           # databases always run as single replica
  selector:
    matchLabels:
      app: books-db     # this deployment manages pods with this label
  template:
    metadata:
      labels:
        app: books-db   # label on the pod — must match selector above
    spec:
      containers:
        - name: books-db
          image: {{ .Values.images.postgres.repository }}:{{ .Values.images.postgres.tag }}
          imagePullPolicy: {{ .Values.images.postgres.pullPolicy }}
          ports:
            - containerPort: 5432
          env:
            - name: POSTGRES_USER
              valueFrom:
                configMapKeyRef:
                  name: app-config          # which ConfigMap
                  key: BOOKS_DB_USER        # which key in that ConfigMap
            - name: POSTGRES_DB
              valueFrom:
                configMapKeyRef:
                  name: app-config
                  key: BOOKS_DB_NAME
            - name: POSTGRES_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: db-secrets          # which Secret
                  key: postgres-password    # which key in that Secret
          resources:
            requests:
              cpu: {{ .Values.resources.postgres.requests.cpu }}
              memory: {{ .Values.resources.postgres.requests.memory }}
            limits:
              cpu: {{ .Values.resources.postgres.limits.cpu }}
              memory: {{ .Values.resources.postgres.limits.memory }}
          volumeMounts:
            - name: books-db-storage
              mountPath: /var/lib/postgresql/data   # where postgres stores data
          # Readiness probe — only send traffic when postgres is ready
          readinessProbe:
            exec:
              command:
                - pg_isready
                - -U
                - postgres
            initialDelaySeconds: 10   # wait 10s before first check
            periodSeconds: 5          # check every 5s
            failureThreshold: 6       # fail 6 times before marking unready
          # Liveness probe — restart if postgres stops responding
          livenessProbe:
            exec:
              command:
                - pg_isready
                - -U
                - postgres
            initialDelaySeconds: 30
            periodSeconds: 10
            failureThreshold: 3
      volumes:
        - name: books-db-storage
          persistentVolumeClaim:
            claimName: books-db-pvc   # uses the PVC defined above

---
# ── Service ───────────────────────────────────────────────────────────
# Gives books-db a stable DNS name inside the cluster
# book-service connects to "books-db:5432" — this Service resolves that
apiVersion: v1
kind: Service
metadata:
  name: books-db        # this name IS the DNS hostname inside the cluster
  namespace: {{ .Values.namespace }}
spec:
  selector:
    app: books-db       # routes traffic to pods with this label
  ports:
    - port: 5432
      targetPort: 5432
  type: ClusterIP       # internal only — not accessible outside cluster
```

---

## templates/users-db.yaml

Identical pattern to books-db, different names and config keys.

```yaml
# helm/book-ordering/templates/users-db.yaml

apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: users-db-pvc
  namespace: {{ .Values.namespace }}
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: {{ .Values.storage.users }}

---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: users-db
  namespace: {{ .Values.namespace }}
  labels:
    app: users-db
spec:
  replicas: 1
  selector:
    matchLabels:
      app: users-db
  template:
    metadata:
      labels:
        app: users-db
    spec:
      containers:
        - name: users-db
          image: {{ .Values.images.postgres.repository }}:{{ .Values.images.postgres.tag }}
          imagePullPolicy: {{ .Values.images.postgres.pullPolicy }}
          ports:
            - containerPort: 5432
          env:
            - name: POSTGRES_USER
              valueFrom:
                configMapKeyRef:
                  name: app-config
                  key: USERS_DB_USER
            - name: POSTGRES_DB
              valueFrom:
                configMapKeyRef:
                  name: app-config
                  key: USERS_DB_NAME
            - name: POSTGRES_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: db-secrets
                  key: postgres-password
          resources:
            requests:
              cpu: {{ .Values.resources.postgres.requests.cpu }}
              memory: {{ .Values.resources.postgres.requests.memory }}
            limits:
              cpu: {{ .Values.resources.postgres.limits.cpu }}
              memory: {{ .Values.resources.postgres.limits.memory }}
          volumeMounts:
            - name: users-db-storage
              mountPath: /var/lib/postgresql/data
          readinessProbe:
            exec:
              command: ["pg_isready", "-U", "postgres"]
            initialDelaySeconds: 10
            periodSeconds: 5
            failureThreshold: 6
          livenessProbe:
            exec:
              command: ["pg_isready", "-U", "postgres"]
            initialDelaySeconds: 30
            periodSeconds: 10
            failureThreshold: 3
      volumes:
        - name: users-db-storage
          persistentVolumeClaim:
            claimName: users-db-pvc

---
apiVersion: v1
kind: Service
metadata:
  name: users-db
  namespace: {{ .Values.namespace }}
spec:
  selector:
    app: users-db
  ports:
    - port: 5432
      targetPort: 5432
  type: ClusterIP
```

---

## templates/orders-db.yaml

```yaml
# helm/book-ordering/templates/orders-db.yaml

apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: orders-db-pvc
  namespace: {{ .Values.namespace }}
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: {{ .Values.storage.orders }}

---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: orders-db
  namespace: {{ .Values.namespace }}
  labels:
    app: orders-db
spec:
  replicas: 1
  selector:
    matchLabels:
      app: orders-db
  template:
    metadata:
      labels:
        app: orders-db
    spec:
      containers:
        - name: orders-db
          image: {{ .Values.images.postgres.repository }}:{{ .Values.images.postgres.tag }}
          imagePullPolicy: {{ .Values.images.postgres.pullPolicy }}
          ports:
            - containerPort: 5432
          env:
            - name: POSTGRES_USER
              valueFrom:
                configMapKeyRef:
                  name: app-config
                  key: ORDERS_DB_USER
            - name: POSTGRES_DB
              valueFrom:
                configMapKeyRef:
                  name: app-config
                  key: ORDERS_DB_NAME
            - name: POSTGRES_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: db-secrets
                  key: postgres-password
          resources:
            requests:
              cpu: {{ .Values.resources.postgres.requests.cpu }}
              memory: {{ .Values.resources.postgres.requests.memory }}
            limits:
              cpu: {{ .Values.resources.postgres.limits.cpu }}
              memory: {{ .Values.resources.postgres.limits.memory }}
          volumeMounts:
            - name: orders-db-storage
              mountPath: /var/lib/postgresql/data
          readinessProbe:
            exec:
              command: ["pg_isready", "-U", "postgres"]
            initialDelaySeconds: 10
            periodSeconds: 5
            failureThreshold: 6
          livenessProbe:
            exec:
              command: ["pg_isready", "-U", "postgres"]
            initialDelaySeconds: 30
            periodSeconds: 10
            failureThreshold: 3
      volumes:
        - name: orders-db-storage
          persistentVolumeClaim:
            claimName: orders-db-pvc

---
apiVersion: v1
kind: Service
metadata:
  name: orders-db
  namespace: {{ .Values.namespace }}
spec:
  selector:
    app: orders-db
  ports:
    - port: 5432
      targetPort: 5432
  type: ClusterIP
```

---

## templates/book-service.yaml

The application deployment. References ConfigMap and Secret for env vars.
Has HTTP-based probes hitting /health endpoint.

```yaml
# helm/book-ordering/templates/book-service.yaml

apiVersion: apps/v1
kind: Deployment
metadata:
  name: book-service
  namespace: {{ .Values.namespace }}
  labels:
    app: book-service
spec:
  replicas: {{ .Values.replicaCount.bookService }}
  selector:
    matchLabels:
      app: book-service
  template:
    metadata:
      labels:
        app: book-service
    spec:
      containers:
        - name: book-service
          image: {{ .Values.images.bookService.repository }}:{{ .Values.images.bookService.tag }}
          imagePullPolicy: {{ .Values.images.bookService.pullPolicy }}
          ports:
            - containerPort: {{ .Values.ports.bookService }}
          # Pull ALL keys from ConfigMap as environment variables
          envFrom:
            - configMapRef:
                name: app-config
          # Pull specific keys from Secret
          env:
            - name: BOOKS_DB_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: db-secrets
                  key: postgres-password
          resources:
            requests:
              cpu: {{ .Values.resources.bookService.requests.cpu }}
              memory: {{ .Values.resources.bookService.requests.memory }}
            limits:
              cpu: {{ .Values.resources.bookService.limits.cpu }}
              memory: {{ .Values.resources.bookService.limits.memory }}
          # Liveness probe — restarts pod if /health fails 3 times
          livenessProbe:
            httpGet:
              path: /health
              port: {{ .Values.ports.bookService }}
            initialDelaySeconds: 15   # give app time to start before checking
            periodSeconds: 10
            failureThreshold: 3
          # Readiness probe — removes pod from load balancer if not ready
          readinessProbe:
            httpGet:
              path: /health
              port: {{ .Values.ports.bookService }}
            initialDelaySeconds: 10
            periodSeconds: 5
            failureThreshold: 3

---
apiVersion: v1
kind: Service
metadata:
  name: book-service      # this is the DNS name order-service uses to call us
  namespace: {{ .Values.namespace }}
spec:
  selector:
    app: book-service
  ports:
    - port: {{ .Values.ports.bookService }}
      targetPort: {{ .Values.ports.bookService }}
  type: ClusterIP
```

**`envFrom` vs `env`** — `envFrom` loads ALL keys from a ConfigMap as env
vars at once. `env` with `valueFrom` loads specific keys — used for secrets
where you want explicit control over which secret values become env vars.

**Why `BOOKS_DB_PASSWORD` separately?** Your Go code reads
`os.Getenv("BOOKS_DB_PASSWORD")` — this env var comes from the Secret,
not the ConfigMap. Sensitive values always come from Secrets, never ConfigMaps.

---

## templates/user-service.yaml

```yaml
# helm/book-ordering/templates/user-service.yaml

apiVersion: apps/v1
kind: Deployment
metadata:
  name: user-service
  namespace: {{ .Values.namespace }}
  labels:
    app: user-service
spec:
  replicas: {{ .Values.replicaCount.userService }}
  selector:
    matchLabels:
      app: user-service
  template:
    metadata:
      labels:
        app: user-service
    spec:
      containers:
        - name: user-service
          image: {{ .Values.images.userService.repository }}:{{ .Values.images.userService.tag }}
          imagePullPolicy: {{ .Values.images.userService.pullPolicy }}
          ports:
            - containerPort: {{ .Values.ports.userService }}
          envFrom:
            - configMapRef:
                name: app-config
          env:
            - name: USERS_DB_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: db-secrets
                  key: postgres-password
            - name: SECRET_KEY
              valueFrom:
                secretKeyRef:
                  name: db-secrets
                  key: jwt-secret
          resources:
            requests:
              cpu: {{ .Values.resources.userService.requests.cpu }}
              memory: {{ .Values.resources.userService.requests.memory }}
            limits:
              cpu: {{ .Values.resources.userService.limits.cpu }}
              memory: {{ .Values.resources.userService.limits.memory }}
          livenessProbe:
            httpGet:
              path: /health
              port: {{ .Values.ports.userService }}
            initialDelaySeconds: 15
            periodSeconds: 10
            failureThreshold: 3
          readinessProbe:
            httpGet:
              path: /health
              port: {{ .Values.ports.userService }}
            initialDelaySeconds: 10
            periodSeconds: 5
            failureThreshold: 3

---
apiVersion: v1
kind: Service
metadata:
  name: user-service
  namespace: {{ .Values.namespace }}
spec:
  selector:
    app: user-service
  ports:
    - port: {{ .Values.ports.userService }}
      targetPort: {{ .Values.ports.userService }}
  type: ClusterIP
```

---

## templates/order-service.yaml

```yaml
# helm/book-ordering/templates/order-service.yaml

apiVersion: apps/v1
kind: Deployment
metadata:
  name: order-service
  namespace: {{ .Values.namespace }}
  labels:
    app: order-service
spec:
  replicas: {{ .Values.replicaCount.orderService }}
  selector:
    matchLabels:
      app: order-service
  template:
    metadata:
      labels:
        app: order-service
    spec:
      containers:
        - name: order-service
          image: {{ .Values.images.orderService.repository }}:{{ .Values.images.orderService.tag }}
          imagePullPolicy: {{ .Values.images.orderService.pullPolicy }}
          ports:
            - containerPort: {{ .Values.ports.orderService }}
          envFrom:
            - configMapRef:
                name: app-config
          env:
            - name: ORDERS_DB_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: db-secrets
                  key: postgres-password
            - name: SECRET_KEY
              valueFrom:
                secretKeyRef:
                  name: db-secrets
                  key: jwt-secret
          resources:
            requests:
              cpu: {{ .Values.resources.orderService.requests.cpu }}
              memory: {{ .Values.resources.orderService.requests.memory }}
            limits:
              cpu: {{ .Values.resources.orderService.limits.cpu }}
              memory: {{ .Values.resources.orderService.limits.memory }}
          livenessProbe:
            httpGet:
              path: /health
              port: {{ .Values.ports.orderService }}
            initialDelaySeconds: 15
            periodSeconds: 10
            failureThreshold: 3
          readinessProbe:
            httpGet:
              path: /health
              port: {{ .Values.ports.orderService }}
            initialDelaySeconds: 10
            periodSeconds: 5
            failureThreshold: 3

---
apiVersion: v1
kind: Service
metadata:
  name: order-service
  namespace: {{ .Values.namespace }}
spec:
  selector:
    app: order-service
  ports:
    - port: {{ .Values.ports.orderService }}
      targetPort: {{ .Values.ports.orderService }}
  type: ClusterIP
```

---

## templates/ingress.yaml

Routes all external traffic to the right service based on path.
Follows the exact same pattern as simple-api ingress on your cluster.

```yaml
# helm/book-ordering/templates/ingress.yaml

apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: book-ordering-ingress
  namespace: {{ .Values.namespace }}
  annotations:
    # tells Kubernetes which ingress controller handles this
    kubernetes.io/ingress.class: nginx
    # strips the path prefix before forwarding to service
    # /books/1 → /books/1 (we keep prefix since our routes include /books)
    nginx.ingress.kubernetes.io/use-regex: "true"
spec:
  ingressClassName: nginx
  rules:
    - host: {{ .Values.ingress.host }}
      http:
        paths:
          - path: /books(/|$)(.*)
            pathType: Prefix
            backend:
              service:
                name: book-service
                port:
                  number: {{ .Values.ports.bookService }}
          - path: /users(/|$)(.*)
            pathType: Prefix
            backend:
              service:
                name: user-service
                port:
                  number: {{ .Values.ports.userService }}
          - path: /login
            pathType: Exact
            backend:
              service:
                name: user-service
                port:
                  number: {{ .Values.ports.userService }}
          - path: /orders(/|$)(.*)
            pathType: Prefix
            backend:
              service:
                name: order-service
                port:
                  number: {{ .Values.ports.orderService }}
```

---

## Deployment Steps

### Step 1 — Copy chart to jump server

On your local machine, zip and copy the helm folder:
```powershell
# zip it
Compress-Archive -Path .\helm -DestinationPath helm.zip

# copy to jump server
scp -i <your-key.pem> helm.zip azureuser@<jump-server-ip>:~/
```

On jump server:
```bash
unzip helm.zip
```

### Step 2 — Verify the chart is valid before deploying

```bash
# lint checks for syntax errors
helm lint ./helm/book-ordering

# dry-run renders all templates and shows what would be applied
# without actually applying anything to the cluster
helm install book-ordering ./helm/book-ordering --dry-run --debug
```

Read the dry-run output — it shows every YAML file that will be created.
Verify env vars, image names, and namespace look correct.

### Step 3 — Deploy

```bash
helm install book-ordering ./helm/book-ordering
```

### Step 4 — Watch it come up

```bash
# watch all pods in the namespace
kubectl get pods -n book-ordering -w

# expected sequence:
# databases start first → Running
# app services start → Init → Running
# all 6 pods Running
```

### Step 5 — Verify everything

```bash
# all pods running
kubectl get pods -n book-ordering

# all services have ClusterIPs
kubectl get svc -n book-ordering

# ingress has an address
kubectl get ingress -n book-ordering

# check logs if any pod fails
kubectl logs -n book-ordering deployment/book-service
kubectl logs -n book-ordering deployment/user-service
kubectl logs -n book-ordering deployment/order-service
```

### Step 6 — Access the app

Since MetalLB/public IP isn't directly accessible, use port-forward
exactly like the simple-api approach:

```bash
# on jump server — forward local port 8080 to ingress controller
kubectl port-forward -n ingress-nginx svc/ingress-nginx-controller 8080:80 &
```

Then test with curl from the jump server:
```bash
# health checks
curl -H "Host: book-ordering.ddns.net" http://localhost:8080/health
# but our health is at service level, so:
curl -H "Host: book-ordering.ddns.net" http://localhost:8080/books

# or if you have DDNS set up pointing to jump server public IP,
# from your local machine:
curl http://book-ordering.ddns.net/books
```

### Step 7 — Upgrading after changes

When you push new images or change values:
```bash
# bump tag in values.yaml then:
helm upgrade book-ordering ./helm/book-ordering

# or override a value without editing values.yaml:
helm upgrade book-ordering ./helm/book-ordering \
  --set images.bookService.tag=v2
```

### Uninstall everything cleanly

```bash
helm uninstall book-ordering
# this removes all resources but keeps PVCs (data safety)
# to also remove PVCs:
kubectl delete pvc -n book-ordering --all
kubectl delete namespace book-ordering
```

---

## One Thing To Check In Your Go Code

Your book-service reads `BOOKS_DB_PASSWORD` but look at database/db.go:
```go
os.Getenv("BOOKS_DB_PASSWORD")
```

Make sure this matches exactly what the Deployment sets as the env var name.
Same for user-service (`USERS_DB_PASSWORD`) and order-service (`ORDERS_DB_PASSWORD`).

If your Go code currently reads just `DB_PASSWORD`, either:
- Update the Go code to read `BOOKS_DB_PASSWORD` 
- Or change the Secret env var name in the deployment to match

They must match exactly — case sensitive.