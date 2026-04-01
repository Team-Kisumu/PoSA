#!/usr/bin/env bash
# PoSA container entrypoint.
# Substitutes the Render $PORT into nginx config and starts all services.

set -e

export PYTHONPATH="/app"

# Render provides $PORT; default to 80 for local Docker runs.
PORT="${PORT:-80}"

# Substitute the port placeholder in the nginx config template.
sed "s/PORT_PLACEHOLDER/${PORT}/g" /etc/nginx/nginx.conf.template > /etc/nginx/nginx.conf

echo "Starting PoSA services..."
echo "  Backend:    :8080"
echo "  AI Engine:  :8000"
echo "  Frontend:   :3000"
echo "  nginx:      :${PORT}"

exec supervisord -c /etc/supervisor/conf.d/posa.conf
