![Gatekeeper Logo](./gatekeeper.png)

## About the project
Gatekeeper is a microservice created to authenticate users for my future apps. It's highly decoupled btw.

## Tech Stack
- **Language:** Go (Golang)
- **Communication:** gRPC & Protocol Buffers (Proto3)
- **Database:** PostgreSQL
- **Caching & Streams:** Redis
- **Auth:** JSON Web Tokens (JWT)
- **Containerization:** Docker & Docker Compose
- **Tools:** Makefile, Evans CLI, Air (Live Reload)

## Features
- [X] **User Registration:** Secure signup with password hashing.
- [X] **Authentication:** JWT-based login system.
- [X] **Password Recovery:** Forgot/Reset password flow using Redis Streams and background workers.
- [X] **Security:** Centralized Auth Interceptor for protected routes.

## What I learned
- **Microservices Architecture:** Gatekeeper itself is a microservice. It's stateless, which means the Go code holds no data; that is PostgreSQL's role, which runs on docker containers. The same happens to Redis.
- **Asynchronous Processing:** Redis Streams were used to send email for recovering a password, just in case the PC running Gatekeeper explodes and the process crashes.
- **Clean Architecture:** Instead of putting all the code in one place, I organized everything in a Clean Architecture®: Handler receives user requests, checking if the request is valid → Service handles the business' rules → Repository receives calls to actually change the data
- **gRPC Deep Dive:** So, I never tried gRPC, it was my first time. Pretty fast and easy to use.
It's fast because of contracts you create with a .proto file. Here, you say what you want to receive and how the method should be like. gRPC itself receives binary data, so no json to slow the process down!
- **Infrastructure as Code:** I found it useful to handle the database and messaging with Postgre and Redis, respectively. It made no sense to have to install a BDOR and Redis in your own system, neither would someone who saw the GitHub repository! So I used docker compose to handle both. Tired of losing time with the terminal every time, Makefile was created.

## Project Structure
```text
├── cmd/                # Entry points (API)
├── internal/
│   ├── handler/        # gRPC implementations
│   ├── service/        # Business logic
│   ├── repository/     # Data access layer
│   ├── interceptor/    # gRPC middlewares (Auth, Rate Limit)
│   ├── model/          # Entities and value objects
│   └── worker/         # Background tasks (Redis consumers)
├── pkg/                # Shared utilities
├── proto/              # Protobuf definitions
├── db/                 # Database migrations
└── compose.yml         # Infrastructure setup
```

## Getting Started

### Prerequisites
- Go 1.21+
- Docker & Docker Compose
- Evans CLI (optional, for testing)

### Installation
1. Clone the repository:
   ```bash
   git clone https://github.com/eduardovfaleiro/gatekeeper.git
   cd gatekeeper
   ```
2. Set up your `.env` file:
   ```bash
   cp .env.example .env
   ```
3. Start the infrastructure:
   ```bash
   make infra
   ```
4. Run migrations:
   ```bash
   make migrateup
   ```
5. Run the application:
   ```bash
   make run
   ```

## Testing the API
You can use **Evans CLI** to interact with the gRPC server:
```bash
make evans
```
Within Evans:
- `package auth`
- `service AuthService`
- `call Register`
