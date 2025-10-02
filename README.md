# 🐋 beia

Simple completions API with rate limiting via Redis. No authentication required.

## Setup

1. Install dependencies:
```bash
go mod download
```

2. Configure environment variables:
```bash
REDIS_URL=redis://localhost:6379
OPENAI_KEY=your_openai_api_key
PORT=8080
```

3. Run:
```bash
go run main.go
```

## Usage

```bash
POST /completions
Content-Type: application/json

{
  "prompt": "Hey!"
}

RESPONSE

{
  "content": "Hey there!"
}
```

## Rate Limits

- 10 requests per minute
- 100 requests per day
- Exceeding minute limit results in 7-day IP ban

## Health Check

```bash
GET /
```
