# TODO-go

# Simple TODO REST API (Go)

A TODO application built with **Go lang** using `HttpServer`.  
Supports basic CRUD operations with **in-memory storage** and **soft delete**.


## Features

- List TODOs (`GET /todos`)
- Create TODO (`POST /todos`)
- Get TODO by ID (`GET /todos/{id}`)
- Update TODO (`PUT /todos/{id}`)
- Delete TODO (`DELETE /todos/{id}`)


## How to Run

```bash
go run main.go
```

server will start at:
http://localhost:8080


Use curl commands for testing:

Get Todo List:
```bash
curl http://localhost:8080/todos
```

Get Todo (by id):
```bash
curl http://localhost:8080/todos/a7f3d354-1c23-497c-ac48-bfe6f4395c95
```

Post Todos:
```bash
curl -X POST http://localhost:8080/todos \
    -H "Content-Type: application/json" \
    -d '{"title":"Test","description":"Test desc","completed":false}'
```

Put Todos:
```bash
curl -X POST http://localhost:8080/todos/a7f3d354-1c23-497c-ac48-bfe6f4395c95 \
  -H "Content-Type: application/json" \
  -d '{"title":"Test","description":"Test desc","completed":true}'
```

Delete Todo (soft delete):
```bash
curl -X DELETE http://localhost:8080/todos/a7f3d354-1c23-497c-ac48-bfe6f4395c95
```
