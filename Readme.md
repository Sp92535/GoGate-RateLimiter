# GoGate Rate Limiter Proxy

## Overview
GoGate is a rate-limiting reverse proxy that controls request flow using various rate-limiting strategies. It forwards allowed requests to their respective destinations while throttling excessive requests. The proxy is highly configurable, allowing per-endpoint and per-method rate limits.

## Prerequisites
- Latest version of Docker and Docker Compose installed
- Latest version of Go installed

## Configuration
Modify the `config/config.yaml` file to define rate limits and routing rules. Below is an example:

```yaml
server:
  host: "localhost"
  port: "6969"

resources:
  - name: Google
    endpoint: /goo
    destination_url: "https://google.com"
    rate_limits:
      GET:
        strategy: LEAKY-BUCKET
        capacity: 10
        rate: 10/s
      POST:
        strategy: TOKEN-BUCKET
        capacity: 10
        rate: 100M/h
        
  - name: Facebook
    endpoint: /face
    destination_url: "https://facebook.com"
    rate_limits:
      GET:
        strategy: SLIDING-WINDOW-LOG
        rate: 10K/s
```

### Supported Rate-Limiting Strategies
- **LEAKY-BUCKET** (requires `capacity`)
- **TOKEN-BUCKET** (requires `capacity`)
- **FIXED-WINDOW**
- **SLIDING-WINDOW**
- **SLIDING-WINDOW-LOG**

### Rate Format Examples
- `10/s` → 10 requests per second
- `10/m` → 10 requests per minute
- `10/h` → 10 requests per hour
- `10M/2m` → 10 million requests per 2 minutes
- `50K/5s` → 50,000 requests per 5 seconds

## Running the Project

### Using Build
```sh
make build
make run
```

### Without Build
```sh
docker compose up -d
go run cmd/main.go
```

## Contributing
Open-source contributions are welcomed! Feel free to fork the repository, create a branch, and submit a pull request with your improvements.