# TODO Go + Postgres + Redis API

A simple TODO REST API built with Go, Postgres, and Redis (for caching). Includes endpoints to create, read, update, patch, and delete TODO items, with cache support and cache-hit indication via HTTP headers.

## Features
- RESTful API for TODO management
- PostgreSQL for persistent storage
- Redis for caching
- Cache-hit responses include `X-Cache-Hit: true` header

## Requirements
- Go 1.18+
- Docker & Docker Compose (recommended for local dev)
- PostgreSQL
- Redis

---

## Running Locally (Terminal, Mac/Linux)

### 1. Clone the repo
```sh
git clone <your-repo-url>
cd TODO-goPostgres
```

### 2. Start with Docker Compose (recommended)
```sh
docker-compose up --build
```
This will start the Go app, Postgres, and Redis using the provided `docker-compose.yml`.

- The API will be available at: `http://localhost:8080`

### 3. Or run manually (requires Go, Postgres, Redis installed)
- Set the `POSTGRES_URI` environment variable appropriately (see Dockerfile for example)
- Start Postgres and Redis services
- Run:
```sh
go run main.go db.go redis.go handler.go model.go
```

---

## Running on Amazon Linux EC2

1. **Install dependencies:**
   ```sh
   sudo yum update -y
   sudo yum install -y git golang docker
   sudo service docker start
   sudo usermod -a -G docker ec2-user
   # Log out and back in for docker group changes to take effect
   ```
2. **(Optional) Install docker-compose:**
   ```sh
   sudo curl -L "https://github.com/docker/compose/releases/download/v2.29.2/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
   sudo chmod +x /usr/local/bin/docker-compose
   ```
3. **Clone and run:**
   ```sh
   git clone <your-repo-url>
   cd TODO-goPostgres
   docker-compose up --build
   ```

- The API will be available on port 8080 (ensure your EC2 security group allows inbound TCP 8080).

---

## API Endpoints
- `GET /todos` - List all todos
- `POST /todos` - Create a todo
- `GET /todos/{id}` - Get a todo by ID
- `PUT /todos/{id}` - Update a todo
- `PATCH /todos/{id}` - Patch a todo
- `DELETE /todos/{id}` - Soft delete a todo

## Notes
- Cache is used for GET endpoints; cache hits are indicated by `X-Cache-Hit: true` header.
- Default Postgres connection string is set in Dockerfile, override with `POSTGRES_URI` if needed.

---

## License
MIT
