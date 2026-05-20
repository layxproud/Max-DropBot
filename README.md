# era-dropbot

A [Max.ru](https://max.ru) messenger bot that accepts files from users and stores them in S3-compatible object storage (MinIO). After a successful upload, the bot sends back a short-lived download link via a built-in URL shortener.

## Features

- Accepts files up to **25 MB** in the following formats:
  - Documents: `.pdf`, `.doc`, `.docx`, `.ppt`, `.pptx`
  - Video: `.mp4`, `.avi`, `.mov`, `.mkv`, `.wmv`
  - Audio: `.mp3`, `.wav`, `.aac`, `.m4a`, `.flac`
  - Images: `.png`, `.jpg`, `.jpeg`, `.gif`, `.tiff`, `.svg`
- Concurrent download worker pool (4 workers)
- Exponential backoff with jitter on transient errors
- Short download URLs with a 24-hour TTL (Redis-backed)
- Structured JSON logging via [zerolog](https://github.com/rs/zerolog)
- Graceful shutdown on `SIGINT` / `SIGTERM`

## Stack

| Component      | Role                      |
| -------------- | ------------------------- |
| Go             | Bot application           |
| MinIO          | Object storage            |
| Redis          | URL shortener token store |
| Nginx          | Reverse proxy             |
| Docker Compose | Orchestration             |

## Getting Started

### Prerequisites

- [Docker](https://docs.docker.com/get-docker/) and [Docker Compose](https://docs.docker.com/compose/)
- A Max.ru bot token

### Configuration

Create a `.env` file in the project root. This file is listed in `.gitignore` and must never be committed.

```dotenv
# Max.ru bot token
BOT_TOKEN=your_bot_token_here

# Redis address (matches the service name in compose.yaml)
REDIS_ADDR=redis:6379

# Public prefix used for generated short links, e.g. https://example.com/f/
SHORT_PREFIX=https://example.com/f/

# MinIO – internal endpoint used for uploads (routed through Nginx)
MINIO_ENDPOINT=nginx:80

# MinIO – public endpoint used for presigned URL generation
MINIO_PUBLIC_ENDPOINT=localhost:9000

# MinIO credentials
MINIO_ROOT_USER=minioadmin
MINIO_ROOT_PASSWORD=minioadmin
```

> **Note:** `MINIO_ENDPOINT` points to the Nginx service so that uploads go through the reverse proxy. `MINIO_PUBLIC_ENDPOINT` is used only for presigned URL signing and does not make network calls.

### Running

```bash
docker compose up --build -d
```

The bot starts polling for updates automatically once MinIO reports healthy.

To stop:

```bash
docker compose down
```

## Project Structure

```
.
├── cmd/bot/          # Entry point
├── internal/
│   ├── attachments/  # File routing by extension
│   ├── bot/          # Update handler, notifier, polling loop
│   ├── downloader/   # Worker pool and download logic
│   ├── messages/     # Bot response strings
│   ├── shortener/    # URL shortener (Redis) + HTTP redirect server
│   └── storage/      # MinIO client wrapper
├── utils/            # Filename sanitization helpers
├── compose.yaml
├── Dockerfile
└── nginx.conf
```

## Bot Commands

| Command  | Description                            |
| -------- | -------------------------------------- |
| `/start` | Welcome message with supported formats |
| `/info`  | List of accepted file formats          |

Any message containing an attachment will trigger the upload flow. The bot responds with a status message once the upload completes or fails.
