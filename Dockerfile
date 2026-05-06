# PoSA Dockerfile
#
# Multi-stage build for Render deployment.
# Single-process architecture: Go backend serves on $PORT,
# proxies /ai/* to the Python AI engine internally.
#
# Build:  docker build -t posa .
# Run:    docker run -p 10000:10000 -e PORT=10000 posa

# ============================================================
# Stage 1: Build Go backend
# ============================================================
FROM golang:1.25-bookworm AS backend-build

WORKDIR /build
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /posa-backend .

# ============================================================
# Stage 2: Build Next.js frontend (static export)
# ============================================================
FROM node:22-bookworm-slim AS frontend-build

WORKDIR /build
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci --ignore-scripts
COPY frontend/ ./

ENV NEXT_PUBLIC_API_URL=""
ENV NEXT_PUBLIC_AI_URL=""

RUN npm run build

# ============================================================
# Stage 3: Production runtime
# ============================================================
FROM node:22-bookworm-slim AS runtime

RUN apt-get update && apt-get install -y --no-install-recommends \
    python3 \
    python3-venv \
    curl \
    nginx \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app
ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1
ENV PYTHONPATH=/app

# --- Go backend binary ---
COPY --from=backend-build /posa-backend /app/posa-backend
RUN chmod +x /app/posa-backend

# --- Python AI engine ---
COPY ai/ /app/ai/
RUN python3 -m venv /app/ai/.venv \
    && /app/ai/.venv/bin/pip install --no-cache-dir -r /app/ai/requirements.txt

# --- Next.js frontend ---
COPY --from=frontend-build /build/.next /app/frontend/.next
COPY --from=frontend-build /build/public /app/frontend/public
COPY --from=frontend-build /build/package.json /app/frontend/package.json
COPY --from=frontend-build /build/node_modules /app/frontend/node_modules

# --- nginx config ---
COPY nginx.conf /app/nginx.conf.template

# Writable paths for nginx.
RUN mkdir -p /tmp/nginx \
    && chmod -R 777 /tmp/nginx /var/log/nginx /var/lib/nginx

# Create start script inline (avoids CRLF/permission issues).
RUN printf '#!/bin/sh\n\
export PYTHONPATH=/app\n\
PORT="${PORT:-10000}"\n\
mkdir -p /tmp/nginx\n\
sed "s/PORT_PLACEHOLDER/${PORT}/g" /app/nginx.conf.template > /tmp/nginx/nginx.conf\n\
echo "Starting services on port ${PORT}..."\n\
/app/posa-backend &\n\
/app/ai/.venv/bin/python -m uvicorn ai.evaluator:app --host 127.0.0.1 --port 8000 --log-level warning &\n\
cd /app/frontend && node node_modules/next/dist/bin/next start -p 3000 &\n\
cd /app\n\
sleep 2\n\
exec nginx -c /tmp/nginx/nginx.conf -g "daemon off;"\n' > /app/start.sh \
    && chmod +x /app/start.sh

EXPOSE 10000

CMD ["/app/start.sh"]
