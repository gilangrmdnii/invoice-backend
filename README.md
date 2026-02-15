# Invoice Backend

REST API backend for project invoice and budget management. Built with Go, Fiber v2, and MySQL.

Provides role-based access control, real-time notifications via Server-Sent Events, audit logging, and a dashboard with aggregated metrics.

---

## Table of Contents

- [Tech Stack](#tech-stack)
- [Architecture](#architecture)
- [Project Structure](#project-structure)
- [Getting Started](#getting-started)
- [Environment Variables](#environment-variables)
- [Database Setup](#database-setup)
- [Running the Server](#running-the-server)
- [Authentication](#authentication)
- [Roles & Permissions](#roles--permissions)
- [API Reference](#api-reference)
  - [Auth](#auth)
  - [Users](#users)
  - [Projects](#projects)
  - [Expenses](#expenses)
  - [Budget Requests](#budget-requests)
  - [Invoices](#invoices)
  - [File Upload](#file-upload)
  - [Dashboard](#dashboard)
  - [Notifications](#notifications)
  - [Audit Logs](#audit-logs)
  - [SSE (Real-time Events)](#sse-real-time-events)
- [Response Format](#response-format)
- [Validation Rules](#validation-rules)
- [Business Rules](#business-rules)
- [Notification Types](#notification-types)
- [Database Schema](#database-schema)

---

## Tech Stack

| Technology | Purpose |
|------------|---------|
| **Go 1.24** | Programming language |
| **Fiber v2** | HTTP framework (built on fasthttp) |
| **MySQL 8.0+** | Relational database |
| **JWT (HS256)** | Stateless authentication |
| **SSE** | Real-time event streaming |
| **bcrypt** | Password hashing |
| **go-playground/validator** | Request input validation |

---

## Architecture

The application follows a **layered architecture** pattern:

```
Request
  |
  v
Middleware  -->  Auth (JWT verification) + Role guard
  |
  v
Handler    -->  Parse HTTP input, call service, return response
  |
  v
Service    -->  Business logic, orchestration, audit & notification
  |
  v
Repository -->  Raw SQL queries to MySQL
  |
  v
Database   -->  MySQL
```

Each layer has a single responsibility:
- **Handler** — HTTP concerns only (parse body, path params, query params, send response)
- **Service** — Business rules, validation, cross-cutting concerns (audit logs, notifications, SSE events)
- **Repository** — Database access via raw SQL with parameterized queries (no ORM)

---

## Project Structure

```
invoice-backend/
├── cmd/
│   └── server/
│       └── main.go                  # Application entrypoint
├── internal/
│   ├── config/
│   │   └── config.go                # Environment configuration loader
│   ├── database/
│   │   └── mysql.go                 # MySQL connection pool setup
│   ├── dto/
│   │   ├── request/                 # Request DTOs with validation tags
│   │   │   ├── auth_request.go
│   │   │   ├── budget_request_request.go
│   │   │   ├── expense_request.go
│   │   │   ├── invoice_request.go
│   │   │   └── project_request.go
│   │   └── response/                # Response DTOs
│   │       ├── audit_log_response.go
│   │       ├── auth_response.go
│   │       ├── budget_request_response.go
│   │       ├── dashboard_response.go
│   │       ├── expense_response.go
│   │       ├── invoice_response.go
│   │       ├── notification_response.go
│   │       └── project_response.go
│   ├── handler/                     # HTTP handlers
│   │   ├── audit_log_handler.go
│   │   ├── auth_handler.go
│   │   ├── budget_request_handler.go
│   │   ├── dashboard_handler.go
│   │   ├── expense_handler.go
│   │   ├── invoice_handler.go
│   │   ├── notification_handler.go
│   │   ├── project_handler.go
│   │   ├── sse_handler.go
│   │   ├── upload_handler.go
│   │   └── user_handler.go
│   ├── middleware/                   # HTTP middleware
│   │   ├── auth.go                  # JWT token verification
│   │   ├── context.go               # Helper to extract user info from context
│   │   └── role.go                  # Role-based access control
│   ├── model/                       # Domain models (maps to DB tables)
│   │   ├── audit_log.go
│   │   ├── budget_request.go
│   │   ├── expense.go
│   │   ├── expense_approval.go
│   │   ├── invoice.go
│   │   ├── notification.go
│   │   ├── project.go
│   │   ├── project_budget.go
│   │   ├── project_member.go
│   │   └── user.go
│   ├── repository/                  # Database access layer (raw SQL)
│   │   ├── audit_log_repository.go
│   │   ├── budget_repository.go
│   │   ├── budget_request_repository.go
│   │   ├── dashboard_repository.go
│   │   ├── expense_repository.go
│   │   ├── helpers.go
│   │   ├── invoice_repository.go
│   │   ├── notification_repository.go
│   │   ├── project_member_repository.go
│   │   ├── project_repository.go
│   │   └── user_repository.go
│   ├── router/
│   │   └── router.go               # Route registration & dependency wiring
│   ├── service/                     # Business logic layer
│   │   ├── auth_service.go
│   │   ├── budget_request_service.go
│   │   ├── dashboard_service.go
│   │   ├── expense_service.go
│   │   ├── invoice_service.go
│   │   ├── notification_service.go
│   │   └── project_service.go
│   └── sse/
│       └── hub.go                   # In-memory SSE pub/sub hub
├── migrations/
│   ├── 000001_init_schema.sql       # Core tables
│   └── 000002_invoices.sql          # Invoice table
├── pkg/
│   ├── jwt/
│   │   └── jwt.go                   # JWT generate & validate
│   ├── response/
│   │   └── response.go              # Standard API response helper
│   └── validator/
│       └── validator.go             # Request body parser & validator
├── uploads/                         # Uploaded files (gitignored)
├── .env.example                     # Environment template
├── .gitignore
├── go.mod
├── go.sum
├── Invoice_Backend.postman_collection.json
└── README.md
```

---

## Getting Started

### Prerequisites

- Go 1.24 or later
- MySQL 8.0 or later

### Installation

```bash
git clone https://github.com/gilangrmdnii/invoice-backend.git
cd invoice-backend
go mod download
```

---

## Environment Variables

Copy the template and fill in your values:

```bash
cp .env.example .env
```

| Variable | Description | Default |
|----------|-------------|---------|
| `APP_PORT` | Server port | `3000` |
| `DB_HOST` | MySQL host | `localhost` |
| `DB_PORT` | MySQL port | `3306` |
| `DB_USER` | MySQL username | `root` |
| `DB_PASSWORD` | MySQL password | _(empty)_ |
| `DB_NAME` | Database name | `invoice_db` |
| `JWT_SECRET` | Secret key for signing JWT tokens | _(required)_ |
| `JWT_EXPIRY_HOURS` | Token expiration time in hours | `24` |

---

## Database Setup

Create the database and run both migration files in order:

```bash
mysql -u root -e "CREATE DATABASE IF NOT EXISTS invoice_db"
mysql -u root invoice_db < migrations/000001_init_schema.sql
mysql -u root invoice_db < migrations/000002_invoices.sql
```

This creates 10 tables: `users`, `projects`, `project_members`, `project_budgets`, `expenses`, `expense_approvals`, `budget_requests`, `invoices`, `notifications`, `audit_logs`.

---

## Running the Server

```bash
go run ./cmd/server/
```

The server starts at `http://localhost:3000` (or the port configured in `APP_PORT`).

### Build & Run Binary

```bash
go build -o server ./cmd/server/
./server
```

---

## Authentication

All endpoints except `/api/auth/*` and `/api/health` require a JWT token.

**How it works:**
1. Register a user via `POST /api/auth/register`
2. Login via `POST /api/auth/login` to receive a JWT token
3. Include the token in all subsequent requests:

```
Authorization: Bearer <token>
```

The token contains the user's `id`, `email`, and `role`, signed with HS256. It expires after the configured `JWT_EXPIRY_HOURS` (default: 24 hours).

---

## Roles & Permissions

| Role | Description |
|------|-------------|
| `SPV` | Supervisor/site personnel. Creates expenses and invoices within assigned projects. Can request additional budget when project budget is depleted. |
| `FINANCE` | Finance team. Manages projects, approves/rejects expenses and budget requests. Views audit logs. |
| `OWNER` | Project owner. Same privileges as FINANCE. |

### Permission Matrix

| Action | SPV | FINANCE | OWNER |
|--------|-----|---------|-------|
| Create project | - | Yes | Yes |
| Update project | - | Yes | Yes |
| Add/remove project members | - | Yes | Yes |
| View projects | Own only | All | All |
| Create expense | Yes | Yes | Yes |
| Approve/reject expense | - | Yes | Yes |
| Create budget request (when budget depleted) | Yes | Yes | Yes |
| Approve/reject budget request | - | Yes | Yes |
| Create invoice | Yes | - | - |
| Update/delete invoice | Own only | - | - |
| View invoices | Own projects | All | All |
| Upload files | Yes | Yes | Yes |
| View dashboard | Own projects | All | All |
| View audit logs | - | Yes | Yes |
| View notifications | Own | Own | Own |
| SSE event stream | Own | Own | Own |

---

## API Reference

**Base URL:** `http://localhost:3000/api`

### Health Check

```
GET /api/health
```

Response:
```json
{ "success": true, "message": "server is running" }
```

---

### Auth

Public endpoints (no token required).

#### Register

```
POST /api/auth/register
```

| Field | Type | Validation |
|-------|------|------------|
| `full_name` | string | Required, 2-255 chars |
| `email` | string | Required, valid email format |
| `password` | string | Required, min 6 chars |
| `role` | string | Required, one of: `SPV`, `FINANCE`, `OWNER` |

Request:
```json
{
  "full_name": "John Doe",
  "email": "john@example.com",
  "password": "secret123",
  "role": "SPV"
}
```

Response `201`:
```json
{
  "success": true,
  "message": "user registered successfully",
  "data": {
    "id": 1,
    "full_name": "John Doe",
    "email": "john@example.com",
    "role": "SPV"
  }
}
```

#### Login

```
POST /api/auth/login
```

| Field | Type | Validation |
|-------|------|------------|
| `email` | string | Required, valid email |
| `password` | string | Required |

Request:
```json
{
  "email": "john@example.com",
  "password": "secret123"
}
```

Response `200`:
```json
{
  "success": true,
  "message": "login successful",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "user": {
      "id": 1,
      "full_name": "John Doe",
      "email": "john@example.com",
      "role": "SPV"
    }
  }
}
```

---

### Users

#### List Users

```
GET /api/users
GET /api/users?role=SPV
```

| Query Param | Type | Description |
|-------------|------|-------------|
| `role` | string | Optional. Filter by role (`SPV`, `FINANCE`, `OWNER`). If omitted, returns all users. |

Response `200`:
```json
{
  "success": true,
  "message": "users retrieved successfully",
  "data": [
    { "id": 1, "full_name": "John Doe", "email": "john@example.com", "role": "SPV", "created_at": "...", "updated_at": "..." }
  ]
}
```

---

### Projects

#### Create Project

```
POST /api/projects
```

**Roles:** FINANCE, OWNER

| Field | Type | Validation |
|-------|------|------------|
| `name` | string | Required, 2-255 chars |
| `description` | string | Optional, max 1000 chars |
| `total_budget` | number | Required, > 0 |

Request:
```json
{
  "name": "Website Redesign",
  "description": "Complete website overhaul",
  "total_budget": 50000000
}
```

A `project_budgets` record is automatically created with the specified `total_budget`.

#### List Projects

```
GET /api/projects
```

SPV sees only projects they are a member of. FINANCE/OWNER sees all.

#### Get Project

```
GET /api/projects/:id
```

#### Update Project

```
PUT /api/projects/:id
```

**Roles:** FINANCE, OWNER

| Field | Type | Validation |
|-------|------|------------|
| `name` | string | Optional, 2-255 chars |
| `description` | string | Optional, max 1000 chars |
| `status` | string | Optional, one of: `ACTIVE`, `COMPLETED`, `ARCHIVED` |

#### Add Member

```
POST /api/projects/:id/members
```

**Roles:** FINANCE, OWNER

| Field | Type | Validation |
|-------|------|------------|
| `user_id` | number | Required |

#### Remove Member

```
DELETE /api/projects/:id/members/:userId
```

**Roles:** FINANCE, OWNER

#### List Members

```
GET /api/projects/:id/members
```

Response includes `full_name` and `email` for each member.

---

### Expenses

#### Create Expense

```
POST /api/expenses
```

| Field | Type | Validation |
|-------|------|------------|
| `project_id` | number | Required |
| `description` | string | Required, 2-1000 chars |
| `amount` | number | Required, > 0 |
| `category` | string | Required, max 255 chars |
| `receipt_url` | string | Optional, max 500 chars |

Request:
```json
{
  "project_id": 1,
  "description": "Office supplies",
  "amount": 500000,
  "category": "Supplies",
  "receipt_url": "/uploads/20260216-abc123.jpg"
}
```

Expense is created with status `PENDING`. Notifications are sent to FINANCE and OWNER users. Audit log is recorded.

#### List Expenses

```
GET /api/expenses
```

SPV sees only expenses from their assigned projects.

#### Get Expense

```
GET /api/expenses/:id
```

#### Update Expense

```
PUT /api/expenses/:id
```

| Field | Type | Validation |
|-------|------|------------|
| `description` | string | Optional, 2-1000 chars |
| `amount` | number | Optional, > 0 |
| `category` | string | Optional, max 255 chars |
| `receipt_url` | string | Optional, max 500 chars |

#### Delete Expense

```
DELETE /api/expenses/:id
```

#### Approve Expense

```
POST /api/expenses/:id/approve
```

**Roles:** FINANCE, OWNER

| Field | Type | Validation |
|-------|------|------------|
| `notes` | string | Optional, max 1000 chars |

Updates expense status to `APPROVED`, adds the amount to the project's `spent_amount`, creates an approval record, sends notification to the expense creator, and records audit log.

#### Reject Expense

```
POST /api/expenses/:id/reject
```

**Roles:** FINANCE, OWNER

| Field | Type | Validation |
|-------|------|------------|
| `notes` | string | Optional, max 1000 chars |

Updates expense status to `REJECTED`, sends notification to the expense creator, and records audit log.

---

### Budget Requests

Budget requests can **only** be created when a project's budget is fully depleted (`spent_amount >= total_budget`). If the budget still has remaining funds, the API returns an error.

#### Create Budget Request

```
POST /api/budget-requests
```

| Field | Type | Validation |
|-------|------|------------|
| `project_id` | number | Required |
| `amount` | number | Required, > 0 |
| `reason` | string | Required, 2-1000 chars |

Request:
```json
{
  "project_id": 1,
  "amount": 10000000,
  "reason": "Budget habis, butuh tambahan untuk material"
}
```

Error if budget not depleted:
```json
{
  "success": false,
  "message": "budget is not yet depleted (remaining: 5000000.00)"
}
```

#### List Budget Requests

```
GET /api/budget-requests
```

SPV sees only requests from their assigned projects.

#### Get Budget Request

```
GET /api/budget-requests/:id
```

#### Approve Budget Request

```
POST /api/budget-requests/:id/approve
```

**Roles:** FINANCE, OWNER

Adds the requested amount to the project's `total_budget`. Sends notification and SSE event to the requester.

#### Reject Budget Request

```
POST /api/budget-requests/:id/reject
```

**Roles:** FINANCE, OWNER

---

### Invoices

Invoices are simple records uploaded by SPV. No approval workflow — they serve as proof of transaction.

#### Create Invoice

```
POST /api/invoices
```

**Roles:** SPV

| Field | Type | Validation |
|-------|------|------------|
| `project_id` | number | Required |
| `amount` | number | Required, > 0 |
| `file_url` | string | Required, max 500 chars |

Request:
```json
{
  "project_id": 1,
  "amount": 5000000,
  "file_url": "/uploads/20260216-invoice.jpg"
}
```

Invoice number is auto-generated in the format `INV-YYYYMMDD-NNNN` (e.g., `INV-20260216-0001`).

The SPV must be a member of the specified project.

#### List Invoices

```
GET /api/invoices
```

SPV sees only invoices from their assigned projects.

#### Get Invoice

```
GET /api/invoices/:id
```

#### Update Invoice

```
PUT /api/invoices/:id
```

**Roles:** SPV (own invoices only)

| Field | Type | Validation |
|-------|------|------------|
| `amount` | number | Optional, > 0 |
| `file_url` | string | Optional, max 500 chars |

#### Delete Invoice

```
DELETE /api/invoices/:id
```

**Roles:** SPV (own invoices only)

---

### File Upload

```
POST /api/upload
```

**Content-Type:** `multipart/form-data`

| Field | Type | Description |
|-------|------|-------------|
| `file` | file | Required. Accepted types: `.jpg`, `.jpeg`, `.png`, `.pdf`. Max size: 5MB. |

Response `200`:
```json
{
  "success": true,
  "message": "file uploaded successfully",
  "data": {
    "file_url": "/uploads/20260216-a3f2e1b4.jpg"
  }
}
```

Files are stored locally in the `./uploads/` directory and served statically at:
```
GET /uploads/<filename>
```

**Typical flow:** Upload file first, then use the returned `file_url` when creating an invoice or expense.

---

### Dashboard

```
GET /api/dashboard
```

Returns aggregated metrics. SPV sees data filtered to their assigned projects. FINANCE/OWNER sees all data.

Response `200`:
```json
{
  "success": true,
  "message": "dashboard retrieved successfully",
  "data": {
    "projects": {
      "total_projects": 5,
      "active_projects": 3
    },
    "budget": {
      "total_budget": 100000000,
      "total_spent": 45000000,
      "remaining": 55000000
    },
    "expenses": {
      "total_expenses": 20,
      "pending_expenses": 5,
      "approved_expenses": 12,
      "rejected_expenses": 3,
      "total_amount": 60000000
    },
    "budget_requests": {
      "total_requests": 8,
      "pending_requests": 2,
      "approved_requests": 5,
      "rejected_requests": 1,
      "total_amount": 50000000
    },
    "invoices": {
      "total_invoices": 10,
      "total_amount": 75000000
    }
  }
}
```

---

### Notifications

Notifications are created automatically when actions occur (expense created, budget approved, etc.) and pushed in real-time via SSE.

#### List Notifications

```
GET /api/notifications
```

Returns all notifications for the authenticated user, sorted by newest first.

#### Get Unread Count

```
GET /api/notifications/unread-count
```

Response:
```json
{
  "success": true,
  "data": { "count": 3 }
}
```

#### Mark Single as Read

```
PATCH /api/notifications/:id/read
```

#### Mark All as Read

```
PATCH /api/notifications/read-all
```

---

### Audit Logs

```
GET /api/audit-logs
GET /api/audit-logs?entity_type=expense
```

**Roles:** FINANCE, OWNER

| Query Param | Type | Description |
|-------------|------|-------------|
| `entity_type` | string | Optional. Filter by entity: `expense`, `budget_request`, `invoice` |

Each log records: who (`user_id`), what (`action`), which entity (`entity_type` + `entity_id`), details (JSON), and when (`created_at`).

---

### SSE (Real-time Events)

```
GET /api/events
```

Opens a persistent Server-Sent Events connection. Events are pushed to the authenticated user in real-time.

**Headers set by server:**
```
Content-Type: text/event-stream
Cache-Control: no-cache
Connection: keep-alive
Transfer-Encoding: chunked
```

**Initial event on connection:**
```
event: connected
data: {"user_id": 1}
```

**Event format:**
```
event: <event_type>
data: {"id": 1, "title": "...", "message": "...", "reference_id": 5}
```

**JavaScript example:**
```javascript
const token = "your-jwt-token";
const es = new EventSource(`http://localhost:3000/api/events?token=${token}`);

// Or use Authorization header via fetch/EventSourcePolyfill
es.addEventListener("expense_created", (e) => {
  const data = JSON.parse(e.data);
  console.log("New expense:", data);
});
```

---

## Response Format

All API responses follow a consistent JSON structure:

**Success:**
```json
{
  "success": true,
  "message": "description of result",
  "data": { }
}
```

**Error:**
```json
{
  "success": false,
  "message": "error description"
}
```

### HTTP Status Codes

| Code | Meaning |
|------|---------|
| `200` | Success |
| `201` | Created |
| `400` | Bad request (validation error) |
| `401` | Unauthorized (missing or invalid token) |
| `403` | Forbidden (insufficient role) |
| `404` | Resource not found |
| `409` | Conflict (e.g., duplicate email) |
| `500` | Internal server error |

---

## Validation Rules

All request bodies are validated using `go-playground/validator`. Common rules:

| Tag | Meaning |
|-----|---------|
| `required` | Field must be present and non-zero |
| `email` | Must be a valid email format |
| `min=N` | Minimum length (string) or value (number) |
| `max=N` | Maximum length |
| `gt=0` | Must be greater than 0 |
| `oneof=A B C` | Must be one of the listed values |
| `omitempty` | Skip validation if field is empty (used for optional update fields) |

Validation error response example:
```json
{
  "success": false,
  "message": "amount must be greater than 0; reason is required"
}
```

---

## Business Rules

1. **Project creation** automatically creates a `project_budgets` record with the specified `total_budget` and `spent_amount = 0`.

2. **Expense approval** adds the expense amount to the project's `spent_amount` in `project_budgets`.

3. **Budget requests** can only be created when the project budget is fully depleted (`spent_amount >= total_budget`). This prevents premature requests.

4. **Budget request approval** adds the requested amount to the project's `total_budget`, effectively increasing available funds.

5. **Invoices** are simple records (no approval workflow). SPV uploads proof of payment; FINANCE/OWNER can view.

6. **SPV data isolation** — SPV users only see data (projects, expenses, invoices, dashboard metrics) from projects they are a member of.

7. **Audit logging** — All create, update, delete, approve, and reject actions on expenses, budget requests, and invoices are recorded in `audit_logs`.

8. **Real-time notifications** — When an action triggers a notification, it is both saved to the database and pushed via SSE to online users.

---

## Notification Types

| Type | Trigger | Recipients |
|------|---------|------------|
| `EXPENSE_CREATED` | New expense submitted | FINANCE, OWNER |
| `EXPENSE_APPROVED` | Expense approved | Expense creator |
| `EXPENSE_REJECTED` | Expense rejected | Expense creator |
| `BUDGET_REQUEST` | New budget request | FINANCE, OWNER |
| `BUDGET_APPROVED` | Budget request approved | Requester |
| `BUDGET_REJECTED` | Budget request rejected | Requester |
| `INVOICE_CREATED` | New invoice uploaded | FINANCE, OWNER |

---

## Database Schema

### Tables

| Table | Description |
|-------|-------------|
| `users` | User accounts with role (`SPV`, `FINANCE`, `OWNER`) |
| `projects` | Projects with status (`ACTIVE`, `COMPLETED`, `ARCHIVED`) |
| `project_members` | Many-to-many: which users belong to which projects |
| `project_budgets` | One-to-one with projects: `total_budget` and `spent_amount` |
| `expenses` | Expense records with status (`PENDING`, `APPROVED`, `REJECTED`) |
| `expense_approvals` | Approval/rejection records for expenses (who, when, notes) |
| `budget_requests` | Requests for additional budget with status |
| `invoices` | Invoice records with auto-generated numbers |
| `notifications` | User notifications with read/unread tracking |
| `audit_logs` | Immutable activity log with JSON details |

### ER Diagram (Simplified)

```
users ──< project_members >── projects ──── project_budgets
  |                              |
  |                              ├──< expenses ──< expense_approvals
  |                              |
  |                              ├──< budget_requests
  |                              |
  |                              └──< invoices
  |
  ├──< notifications
  └──< audit_logs
```

Full schema definitions are in the [`migrations/`](migrations/) directory.

---

## Postman Collection

A ready-to-use Postman collection is included:

```
Invoice_Backend.postman_collection.json
```

**Features:**
- All 37+ endpoints organized in folders
- Example request bodies for every endpoint
- Auto-save JWT token on login (via test script)
- Collection-level Bearer auth (no need to set token per request)
- Base URL configurable via `{{base_url}}` variable

**Import:** Open Postman > Import > select the JSON file.
