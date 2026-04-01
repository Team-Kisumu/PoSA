#!/bin/sh
# PoSA container entrypoint.
# Starts all services. Works as non-root (Render requirement).

set -e

export PYTHONPATH="/app"
PORT="${PORT:-10000}"

# Generate nginx config with the correct port in a writable location.
mkdir -p /tmp/nginx
sed "s/PORT_PLACEHOLDER/${PORT}/g" /app/nginx.conf.template > /tmp/nginx/nginx.conf

echo "Starting PoSA services..."

# Start backend.
/app/backend/posa-backend &

# Start AI engine.
/app/ai/.venv/bin/python -m uvicorn ai.evaluator:app --host 0.0.0.0 --port 8000 &

# Start frontend.
cd /app/frontend && node_modules/.bin/next start -p 3000 &
cd /app

echo "  Backend:   :8080"
echo "  AI Engine: :8000"
echo "  Frontend:  :3000"
echo "  nginx:     :${PORT}"

# Start nginx in foreground with the generated config.
exec nginx -c /tmp/nginx/nginx.conf -g "daemon off;"
