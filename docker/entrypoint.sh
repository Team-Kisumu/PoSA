#!/usr/bin/env bash
# PoSA container entrypoint.
# Loads environment variables and starts all services via supervisor.

set -e

# Export all env vars so child processes (backend, AI engine) inherit them.
# Docker --env-file and -e flags set vars in this process; supervisor
# passes them to child processes via the environment.
export PYTHONPATH="/app"

# Install supervisor if not present (lightweight process manager).
if ! command -v supervisord &> /dev/null; then
    pip3 install --no-cache-dir supervisor
fi

echo "Starting PoSA services..."
echo "  Backend:    :8080"
echo "  AI Engine:  :8000"
echo "  Frontend:   :3000"
echo "  nginx:      :80"

exec supervisord -c /etc/supervisor/conf.d/posa.conf
