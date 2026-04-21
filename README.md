docker-compose.yml for ubuntu

services:
  bot:
    build: .
    container_name: my-bot
    restart: always
    env_file: ".env"
    depends_on:
      minio:
        condition: service_healthy
      redis:
        condition: service_started

  minio:
    image: minio/minio
    container_name: minio
    command: server /data --console-address ":9001"
    env_file: ".env"
    ports:
      - "9000:9000"
      - "9001:9001"
    volumes:
      - minio-data:/data
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:9000/minio/health/live"]
      interval: 5s
      retries: 5
    
  redis:
    image: redis:7-alpine
    restart: always
    volumes:
      - redis-data:/data
    command: redis-server --save 60 1

  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf:ro
    depends_on:
      - bot

volumes:
  minio-data:
  redis-data:

nginx.conf for ubuntu

events {}
http {
  server_tokens off;

  server {
    listen 80;
    server_name 10.13.1.194;

    location /f/ {
      proxy_pass       http://bot:8080;
      proxy_set_header Host $host;
      proxy_set_header X-Real-IP $remote_addr;
    }

    location / {
      return 404;
    }
  }
}
