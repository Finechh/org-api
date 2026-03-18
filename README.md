# org-api

![Go](https://img.shields.io/badge/Go-1.25-00ADD8?style=flat&logo=go&logoColor=white)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-4169E1?style=flat&logo=postgresql&logoColor=white)
![GORM](https://img.shields.io/badge/GORM-1.25-00ACD7?style=flat)
![Docker](https://img.shields.io/badge/Docker-compose-2496ED?style=flat&logo=docker&logoColor=white)
![Coverage](https://img.shields.io/badge/Coverage-69.6%25-yellow?style=flat)

REST API for managing organizational structure — departments and employees.

---

## Run

```bash
docker-compose up --build
```

API available at `http://localhost:8080`. Migrations apply automatically on startup.

To reset the database:

```bash
docker-compose up --build
```

---

## Tests

```bash
go test ./internal/handler/... -v
```

With coverage:

```bash
go test ./internal/handler/... -cover
```

---

## Endpoints

| Method | Path                           | Description                  |
| ------ | ------------------------------ | ---------------------------- |
| POST   | `/departments/`                | Create department            |
| GET    | `/departments/{id}`            | Get department with tree     |
| PATCH  | `/departments/{id}`            | Update name / parent         |
| DELETE | `/departments/{id}`            | Delete (cascade or reassign) |
| POST   | `/departments/{id}/employees/` | Add employee                 |

### GET query params

| Param               | Type | Default | Description                         |
| ------------------- | ---- | ------- | ----------------------------------- |
| `depth`             | int  | 1       | Depth of nested departments (max 5) |
| `include_employees` | bool | true    | Include employees in response       |

### DELETE query params

| Param                       | Values                 | Description                                               |
| --------------------------- | ---------------------- | --------------------------------------------------------- |
| `mode`                      | `cascade` / `reassign` | Delete mode                                               |
| `reassign_to_department_id` | int                    | Target department for employees (required for `reassign`) |

---

## Examples

```bash
# Create department
curl -X POST http://localhost:8080/departments/ \
  -H "Content-Type: application/json" \
  -d '{"name": "Engineering"}'

# Create child department
curl -X POST http://localhost:8080/departments/ \
  -H "Content-Type: application/json" \
  -d '{"name": "Backend", "parent_id": 1}'

# Get department tree
curl "http://localhost:8080/departments/1?depth=3&include_employees=true"

# Add employee
curl -X POST http://localhost:8080/departments/1/employees/ \
  -H "Content-Type: application/json" \
  -d '{"full_name": "Ivan Ivanov", "position": "Senior Engineer", "hired_at": "2023-01-15"}'

# Move department
curl -X PATCH http://localhost:8080/departments/2 \
  -H "Content-Type: application/json" \
  -d '{"parent_id": 3}'

# Delete cascade
curl -X DELETE "http://localhost:8080/departments/2?mode=cascade"

# Delete with employee reassignment
curl -X DELETE "http://localhost:8080/departments/2?mode=reassign&reassign_to_department_id=1"
```

---

## Project structure

```
├── cmd/api/          # entrypoint
├── config/           # env config
├── internal/
│   ├── app/          # app container
│   ├── database/     # db connection and migrations
│   ├── handler/      # HTTP handlers + tests
│   ├── middleware/   # logging, recover
│   ├── model/        # models and DTOs
│   ├── repository/   # db layer (GORM)
│   └── service/      # business logic
├── pkg/logger/       # logger
├── Dockerfile
└── docker-compose.yml
```

---

## Environment variables

| Variable      | Default   | Description       |
| ------------- | --------- | ----------------- |
| `DB_HOST`     | localhost | PostgreSQL host   |
| `DB_PORT`     | 5432      | PostgreSQL port   |
| `DB_USER`     | postgres  | Database user     |
| `DB_PASSWORD` | postgres  | Database password |
| `DB_NAME`     | orgapi    | Database name     |
| `SERVER_PORT` | 8080      | HTTP server port  |
