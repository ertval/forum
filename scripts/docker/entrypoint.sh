#!/bin/sh
set -eu

if [ "$#" -eq 0 ]; then
  set -- ./forum
fi

if [ "$(id -u)" = "0" ]; then
  mkdir -p /app/data /app/static/uploads
  chown -R appuser:appuser /app/data /app/static/uploads || true
  exec su-exec appuser "$@"
fi

exec "$@"
