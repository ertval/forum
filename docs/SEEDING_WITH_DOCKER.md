# Seeding and Testing Without Local sqlite3

This project can be seeded and tested using Docker only.

## Recommended workflow

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

## Efficient Docker usage

- Keep `./data:/app/data` for DB persistence.
- Keep `./static/uploads:/app/static/uploads` for image persistence.
- Use `docker compose logs -f forum` during script runs.
- Use disposable tools containers for sqlite/go toolchain, not the runtime app container.

## Optional dev container approach

Use a development container with Go + sqlite3 + CGO toolchain so host setup stays minimal.
Recommended starter files:
- `.devcontainer/devcontainer.json`
- `.devcontainer/Dockerfile`

This is useful for Windows machines where local sqlite/cgo setup is inconsistent.