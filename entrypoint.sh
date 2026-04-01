#!/bin/sh
# PoSA container entrypoint.
# Starts all services and waits for any to exit.

set -e

export PYTHONPATH="/app"
PORT="${PORT:-80}"

# Generate nginx config from template with the correct port.
sed "s/PORT_PLACEHOLDER/${PORT}/g" /etc/nginx/nginx.conf.template > /etc/nginx/nginx.conf

echo "Starting PoSA services..."

# Start backend.
/app/backend/posa-backend &

# Start AI engine.
/app/ai/.venv/bin/python -m uvicorn ai.evaluator:app --host 0.0.0.0 --port 8000 &

# Start frontend.
cd /app/frontend && node_modules/.bin/next start -p 3000 &
cd /app

# Start nginx in foreground (keeps container alive).
echo "  Backend:   :8080"
echo "  AI Engine: :8000"
echo "  Frontend:  :3000"
echo "  nginx:     :${PORT}"

exec nginx -g "daemon off;"
