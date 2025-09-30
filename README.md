# WarehouseControl

WarehouseControl is a mini inventory management system with CRUD functionality, role-based access control, and full change history logging.

In a warehouse, it’s important to know **who did what and when**. Different employees have different permissions: a warehouse worker can edit, a manager can only view, and an auditor can see the full history.

This project intentionally uses **database triggers for logging** as an anti-pattern, to demonstrate why this approach is generally not recommended in real-world projects.

---

## Features

* CRUD operations for inventory items:

    * `POST /api/items` — create a new item
    * `GET /api/items` — list all items
    * `GET /api/items/{id}` — get a specific item
    * `PUT /api/items/{id}` — update an item
    * `DELETE /api/items/{id}` — delete an item
* Full change history (who, when, what changed) stored in the database
* Role-based access control:

    * `admin` — full access
    * `manager` — view and edit
    * `viewer` — read-only
* JWT-based authentication with role validation
* Simple web UI for:

    * Login as a user with a specific role
    * View, add, edit, delete inventory items (depending on permissions)
    * View item change history
    * Compare item versions

---

## API Endpoints

### Auth

* `POST /api/auth/register` — register a new user
* `POST /api/auth/login` — login and receive JWT token

### Users

* `GET /api/users/{id}` — get user info (protected, requires JWT)

### Items

* `GET /api/items` — list items (public)
* `GET /api/items/{id}` — get item details (public)
* `POST /api/items` — create item (admin, manager)
* `PUT /api/items/{id}` — update item (admin, manager)
* `DELETE /api/items/{id}` — delete item (admin)

### Audit

* `GET /api/audit/items/{id}/history` — get item change history (admin)
* `POST /api/audit/items/compare` — compare two item versions (admin)

---

## Project Architecture

```
.
├── cmd/                 # Application entry points
├── config/              # Configuration files
├── internal/            # Internal application packages
│   ├── api/             # HTTP handlers, router, server
│   │   ├── request      # Request parsing helpers (UUID, float, time etc.)
│   │   ├── response     # Response helpers (JSON, OK, Created, Fail)
│   │   ├── router       # Route definitions
│   │   └── server       # HTTP server initialization
│   ├── config/          # Config parsing logic
│   ├── middleware/      # JWT auth, role-based middleware
│   ├── model/           # Data models (Item, User, ItemHistory etc.)
│   ├── repository/      # Database repository layer
│   └── service/         # Business logic
├── migrations/          # Database migrations
├── web/                 # Frontend UI (HTML/JS, or React/TS/TailwindCSS)
├── Dockerfile           # Backend Dockerfile
├── go.mod
├── go.sum
├── .env.example         # Example environment variables
├── docker-compose.yml   # Multi-service Docker setup
├── Makefile             # Development commands
└── README.md
```

---

## Running the Project

### Using Makefile

* Build and start all Docker services:

```bash
make docker-up
```

* Stop and remove all Docker services and volumes:

```bash
make docker-down
```

---

### Default Ports

* Frontend: [http://localhost:3000](http://localhost:3000)
* Backend API: [http://localhost:8080/api](http://localhost:8080/api)

---

### Notes

* Item history is stored via **database triggers** (anti-pattern) for learning purposes.
* JWT tokens carry the role of the user for access control.
* The frontend UI allows testing all roles and operations directly.