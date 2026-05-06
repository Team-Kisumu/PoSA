#!/bin/sh
# PoSA container entrypoint for Render.
# Starts backend, AI engine, frontend, then nginx on $PORT.

export PYTHONPATH="/app"
PORT="${PORT:-10000}"

# Generate nginx config.
mkdir -p /tmp/nginx
sed "s/PORT_PLACEHOLDER/${PORT}/g" /app/nginx.conf.template > /tmp/nginx/nginx.conf

# Trap signals to clean up child processes.
cleanup() {
    kill $(jobs -p) 2>/dev/null
    exit 0
}
trap cleanup TERM INT

# Start services.
/app/posa-backend &
/app/ai/.venv/bin/python -m uvicorn ai.evaluator:app --host 127.0.0.1 --port 8000 --log-level warning &
cd /app/frontend && npx next start -p 3000 &
cd /app

# Wait for backend to be ready before starting nginx.
for i in 1 2 3 4 5 6 7 8 9 10; do
    if curl -sf http://127.0.0.1:8080/health > /dev/null 2>&1; then
        break
    fi
    sleep 1
done

echo "PoSA ready on port ${PORT}"

# Run nginx in foreground.
nginx -c /tmp/nginx/nginx.conf -g "daemon off;" &
NGINX_PID=$!

# Wait for any process to exit.
wait $NGINX_PID
