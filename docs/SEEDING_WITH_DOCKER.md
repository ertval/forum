# Seeding and Testing Without Local sqlite3

This project can be seeded and tested using Docker only.

## Recommended workflow (Docker Compose)

Start app stack:

```bash
docker compose up -d --build
```

Run seed script in a disposable tools container:

```bash
docker run --rm -v "${PWD}:/workspace" -w /workspace alpine:3.20 sh -lc \
  "apk add --no-cache bash sqlite openssl >/dev/null && DATABASE_PATH=/workspace/data/forum.db bash scripts/seed/seed.sh"
```

Reset and reseed:

```bash
docker compose down
rm -f data/forum.db
docker run --rm -v "${PWD}:/workspace" -w /workspace alpine:3.20 sh -lc \
  "apk add --no-cache bash sqlite openssl >/dev/null && DATABASE_PATH=/workspace/data/forum.db bash scripts/seed/seed.sh"
docker compose up -d
```

## Standalone Docker usage (without Compose)

```bash
# Build the image
docker build -t forum .

# First run — create a named container with persistent volumes
docker run -d --name forum -p 8080:8080 \
  -v forum-data:/app/data \
  -v forum-uploads:/app/static/uploads \
  forum

# Stop the container
docker stop forum

# Restart (reuses the same container and data)
docker start forum

# View live logs
docker logs -f forum

# Remove the container (data volumes are preserved)
docker rm forum

# Remove data volumes too if you want a clean slate
docker volume rm forum-data forum-uploads
```

The app binds to `0.0.0.0` by default so no extra environment variables are needed.
To seed a standalone container, mount its data volume into the tools container:

```bash
docker run --rm -v forum-data:/workspace/data -w /workspace alpine:3.20 sh -lc \
  "apk add --no-cache bash sqlite openssl >/dev/null && DATABASE_PATH=/workspace/data/forum.db bash scripts/seed/seed.sh"
```

## Efficient Docker usage

- Keep `./data:/app/data` for DB persistence (Compose) or use named volumes (standalone).
- Keep `./static/uploads:/app/static/uploads` for image persistence.
- Use `docker compose logs -f forum` or `docker logs -f forum` during script runs.
- Use disposable tools containers for sqlite/go toolchain, not the runtime app container.

## Optional dev container approach

Use a development container with Go + sqlite3 + CGO toolchain so host setup stays minimal.
Recommended starter files:
- `.devcontainer/devcontainer.json`
- `.devcontainer/Dockerfile`

This is useful for Windows machines where local sqlite/cgo setup is inconsistent.