Digital-Wallet
Digital Wallet Platform Backend

A high-performance backend system for a digital wallet platform that handles millions of transactions, featuring session management, caching, webhook processing, reporting, and more.
🚀 Features

    Session Management: JWT-based authentication with Redis-backed sessions

    Caching: Redis caching with write-behind strategy for optimal performance

    Webhook Handling: Secure processing of webhook events with retry logic

    Bulk Data Generation: Tool for simulating large datasets for testing

    Dynamic Fee Calculation: Flexible fee engine with tier-based rules

    Tier Management: Role-based access control for different user tiers

    Reporting Engine: Comprehensive transaction reporting with CSV export

    Observability: Integrated logging and metrics

🏗️ Architecture

Architecture Diagram
🛠️ Getting Started
Prerequisites

    Go 1.21+

    Redis 7+

    PostgreSQL 15+

    Docker (optional, for containerized environments)

Installation

    Clone the repository:

git clone https://github.com/digital-wallet.git
cd digital-wallet

Set up environment variables:

cp .env.example .env
# Edit .env with your configuration

Run with Docker:

docker-compose up -d

Or run locally:

    go run cmd/main.go

📘 API Documentation
🔐 Authentication

Login

POST /api/login

🔁 Webhooks

Process Webhook

POST /api/webhook/notify

🧪 Simulation

Generate Bulk Data

POST /api/simulate/bulk

Check Simulation Status

GET /api/simulate/status

📊 Reporting

Get Transaction Report

GET /api/reports/transactions

⚙️ Configuration
Environment Variable	Description	Default
ENV	Application environment	development
REDIS_URL	Redis connection URL	redis://localhost:6379
DB_URL	PostgreSQL connection URL	postgres://user:pass@localhost:5432/wallet
JWT_SECRET	Secret for JWT signing	-
WEBHOOK_SECRET	Secret for webhook HMAC validation	-
🧪 Testing

Run unit tests:

go test ./...

Run integration tests (requires Redis and PostgreSQL):

go test -tags=integration ./...

Load testing with k6:

k6 run loadtest/transactions.js

🚢 Deployment

The service is designed to run in a containerized environment. A sample Kubernetes deployment is available in:

deploy/k8s/

📈 Monitoring

    Prometheus metrics available at:

    /metrics

    Integrated with Loki for centralized logging

🤝 Contributing

    Fork the repository

    Create a feature branch

    Commit your changes

    Push to the branch

    Create a Pull Request

📄 License

MIT
