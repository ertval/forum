FROM mcr.microsoft.com/devcontainers/go:1-1.24-bookworm

RUN apt-get update && apt-get install -y \
    sqlite3 \
    libsqlite3-dev \
    build-essential \
    openssl \
    curl \
    jq \
    lsof \
    procps \
 && rm -rf /var/lib/apt/lists/*