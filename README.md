# Shortener — URL Shortening Service with Analytics

Shortener is a mini URL shortening service. It allows you to generate short links, redirect users to the original URLs, and collect analytics on link usage (who clicked, when, and from which device).  

---

## Features

- **Shorten URLs**: Generate short URLs for any long link.  
- **Redirect**: Redirect users from short URLs to original URLs.  
- **Analytics**: Track visits including:
  - Total clicks
  - Clicks per day
  - Clicks by User-Agent / device
  - Cache popular links using Redis
  - Custom aliases
  - Simple frontend UI to test the service

---

## Project Structure

```

.
├── backend/                 # Backend service
│   ├── cmd/                 # Application entry points
│   ├── config/              # Configuration files
│   ├── internal/            # Internal application packages
│   │   ├── api/             # HTTP handlers, router, server
│   │   ├── config/          # Config parsing logic
│   │   ├── middlewares/     # HTTP middlewares
│   │   ├── model/           # Data models
│   │   ├── repository/      # Database repositories
│   │   ├── service/         # Business logic
│   ├── migrations/          # Database migrations
│   ├── Dockerfile           # Backend Dockerfile
│   ├── go.mod
│   └── go.sum
├── frontend/                # Frontend application
├── .env.example             # Example environment variables
├── docker-compose.yml       # Multi-service Docker setup
├── Makefile                 # Development commands
└── README.md

````
---

## Running the Project
Copy and update .env:

```
cp .env.example .env
```

Build and run services via Docker:

```
make docker-up
```

The backend will be available at:

- Backend API → http://localhost:8080/api/notify
- Frontend UI → http://localhost:3000

To stop services:

```
make docker-down
```

---

## API Endpoints

| Method | Endpoint                | Description                        |
| ------ | ----------------------- | ---------------------------------- |
| POST   | `/api/shorten`          | Create a new short URL             |
| GET    | `/api/s/:alias`         | Redirect to the original URL       |
| GET    | `/api/analytics/:alias` | Retrieve analytics for a short URL |

---

## Example Requests

### **1. Create Short URL**

**Request**

```http
POST /api/shorten
Content-Type: application/json

{
  "url": "https://example.com/long-url",
  "alias": "my-short-link"   // optional
}
```

**Response**

```json
{
  "url": "https://example.com/long-url",
  "alias": "my-short-link",
  "created_at": "2025-09-18T12:00:00Z"
}
```

---

### **2. Redirect Short URL**

Access via browser or HTTP client:

```
GET /api/s/my-short-link
```

Redirects to: `https://example.com/long-url`

---

### **3. Get Analytics**

**Request**

```
GET /api/analytics/my-short-link
```

**Response**

```json
{
  "alias": "my-short-link",
  "total_clicks": 42,
  "daily": {
    "2025-09-16": 5,
    "2025-09-17": 10,
    "2025-09-18": 27
  },
  "user_agent": {
    "Chrome": 30,
    "Firefox": 12
  }
}
```

## Summary
- Backend (Go + PostgreSQL + Redis) → runs on port 8080
- Frontend → runs on port 3000
- URL can be shorten via API or UI
