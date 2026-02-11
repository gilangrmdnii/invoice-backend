# Invoice Backend

REST API backend for project invoice management built with Go, Fiber, and MySQL.

## Tech Stack

- **Go 1.24** — language
- **Fiber v2** — HTTP framework
- **MySQL** — database
- **JWT** — authentication
- **SSE** — real-time notifications

## Getting Started

### Prerequisites

- Go 1.24+
- MySQL 8.0+

### Setup

1. Clone the repository:

```bash
git clone https://github.com/gilangrmdnii/invoice-backend.git
cd invoice-backend
```

2. Copy environment file and configure:

```bash
cp .env.example .env
```

```
APP_PORT=3000
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=
DB_NAME=invoice_db
JWT_SECRET=your-secret-key-here
JWT_EXPIRY_HOURS=24
```

3. Create the database and run migrations:

```bash
mysql -u root -e "CREATE DATABASE IF NOT EXISTS invoice_db"
mysql -u root invoice_db < migrations/000001_init_schema.sql
```

4. Install dependencies and run:

```bash
go mod download
go run ./cmd/server/
```

Server starts at `http://localhost:3000`.

## Project Structure

```
├── cmd/server/             # Application entrypoint
├── internal/
│   ├── config/             # Environment configuration
│   ├── database/           # MySQL connection
│   ├── dto/
│   │   ├── request/        # Request validation structs
│   │   └── response/       # Response DTOs
│   ├── handler/            # HTTP handlers
│   ├── middleware/          # Auth & role middleware
│   ├── model/              # Domain models
│   ├── repository/         # Database queries
│   ├── router/             # Route registration
│   ├── service/            # Business logic
│   └── sse/                # Server-Sent Events hub
├── migrations/             # SQL migration files
└── pkg/
    ├── jwt/                # JWT helper
    ├── response/           # Standard API response
    └── validator/          # Request validator
```

## Roles

| Role | Description |
|------|-------------|
| `SPV` | Supervisor — manages expenses within assigned projects |
| `FINANCE` | Finance — approves/rejects expenses and budget requests, full visibility |
| `OWNER` | Owner — same privileges as FINANCE |

## API Endpoints

### Auth (Public)

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/auth/register` | Register a new user |
| POST | `/api/auth/login` | Login and get JWT token |

**Register:**
```json
{
  "full_name": "John Doe",
  "email": "john@example.com",
  "password": "secret123",
  "role": "SPV"
}
```

**Login:**
```json
{
  "email": "john@example.com",
  "password": "secret123"
}
```

All protected endpoints require the header:
```
Authorization: Bearer <token>
```

### Projects

| Method | Path | Role Guard | Description |
|--------|------|------------|-------------|
| POST | `/api/projects` | FINANCE, OWNER | Create project |
| GET | `/api/projects` | All | List projects (SPV: own projects only) |
| GET | `/api/projects/:id` | All | Get project by ID |
| PUT | `/api/projects/:id` | FINANCE, OWNER | Update project |
| POST | `/api/projects/:id/members` | FINANCE, OWNER | Add member |
| DELETE | `/api/projects/:id/members/:userId` | FINANCE, OWNER | Remove member |
| GET | `/api/projects/:id/members` | All | List members |

### Expenses

| Method | Path | Role Guard | Description |
|--------|------|------------|-------------|
| POST | `/api/expenses` | All | Create expense |
| GET | `/api/expenses` | All | List expenses (SPV: own projects only) |
| GET | `/api/expenses/:id` | All | Get expense by ID |
| PUT | `/api/expenses/:id` | All | Update expense (PENDING only) |
| DELETE | `/api/expenses/:id` | All | Delete expense (PENDING only) |
| POST | `/api/expenses/:id/approve` | FINANCE, OWNER | Approve expense |
| POST | `/api/expenses/:id/reject` | FINANCE, OWNER | Reject expense |

### Budget Requests

| Method | Path | Role Guard | Description |
|--------|------|------------|-------------|
| POST | `/api/budget-requests` | All | Create budget request |
| GET | `/api/budget-requests` | All | List budget requests |
| GET | `/api/budget-requests/:id` | All | Get budget request by ID |
| POST | `/api/budget-requests/:id/approve` | FINANCE, OWNER | Approve request |
| POST | `/api/budget-requests/:id/reject` | FINANCE, OWNER | Reject request |

### Dashboard

| Method | Path | Role Guard | Description |
|--------|------|------------|-------------|
| GET | `/api/dashboard` | All | Aggregated stats (SPV: filtered by own projects) |

**Response:**
```json
{
  "success": true,
  "message": "dashboard retrieved successfully",
  "data": {
    "projects": { "total_projects": 5, "active_projects": 3 },
    "budget": { "total_budget": 100000, "total_spent": 45000, "remaining": 55000 },
    "expenses": { "total_expenses": 20, "pending_expenses": 5, "approved_expenses": 12, "rejected_expenses": 3, "total_amount": 60000 },
    "budget_requests": { "total_requests": 8, "pending_requests": 2, "approved_requests": 5, "rejected_requests": 1, "total_amount": 50000 }
  }
}
```

### Notifications

| Method | Path | Role Guard | Description |
|--------|------|------------|-------------|
| GET | `/api/notifications` | All | List user's notifications |
| GET | `/api/notifications/unread-count` | All | Get unread count |
| PATCH | `/api/notifications/read-all` | All | Mark all as read |
| PATCH | `/api/notifications/:id/read` | All | Mark one as read |

### Audit Logs

| Method | Path | Role Guard | Description |
|--------|------|------------|-------------|
| GET | `/api/audit-logs` | FINANCE, OWNER | List audit logs |
| GET | `/api/audit-logs?entity_type=expense` | FINANCE, OWNER | Filter by entity type |

### SSE (Server-Sent Events)

| Method | Path | Role Guard | Description |
|--------|------|------------|-------------|
| GET | `/api/events` | All | Real-time event stream |

Connect with any SSE client:
```bash
curl -N -H "Authorization: Bearer <token>" http://localhost:3000/api/events
```

Events are pushed when:
- An expense is created, approved, or rejected
- A budget request is created, approved, or rejected

## API Response Format

All endpoints return a standard JSON format:

```json
{
  "success": true,
  "message": "description of result",
  "data": {}
}
```

## Notification Types

| Type | Trigger | Recipients |
|------|---------|------------|
| `EXPENSE_CREATED` | New expense submitted | FINANCE, OWNER |
| `EXPENSE_APPROVED` | Expense approved | Expense creator |
| `EXPENSE_REJECTED` | Expense rejected | Expense creator |
| `BUDGET_REQUEST` | New budget request submitted | FINANCE, OWNER |
| `BUDGET_APPROVED` | Budget request approved | Requester |
| `BUDGET_REJECTED` | Budget request rejected | Requester |

## Database Schema

Tables: `users`, `projects`, `project_members`, `project_budgets`, `expenses`, `expense_approvals`, `budget_requests`, `notifications`, `audit_logs`.

See [`migrations/000001_init_schema.sql`](migrations/000001_init_schema.sql) for the full schema.
