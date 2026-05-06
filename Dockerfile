# PoSA Dockerfile
#
# Multi-stage build for Render deployment.
# All services run behind nginx on $PORT.
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
RUN CGO_ENABLED=0 GOOS=linux go build -o /posa-backend .

# ============================================================
# Stage 2: Build Next.js frontend
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
    nginx \
    python3 \
    python3-venv \
    curl \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app
ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1
ENV PYTHONPATH=/app

# --- Go backend binary ---
COPY --from=backend-build /posa-backend /app/posa-backend

# --- Python AI engine ---
COPY ai/ /app/ai/
RUN python3 -m venv /app/ai/.venv \
    && /app/ai/.venv/bin/pip install --no-cache-dir -r /app/ai/requirements.txt

# --- Next.js frontend ---
COPY --from=frontend-build /build/.next /app/frontend/.next
COPY --from=frontend-build /build/public /app/frontend/public
COPY --from=frontend-build /build/package.json /app/frontend/package.json
COPY --from=frontend-build /build/node_modules /app/frontend/node_modules

# --- nginx config template ---
COPY nginx.conf /app/nginx.conf.template

# Writable paths for non-root nginx.
RUN mkdir -p /tmp/nginx \
    && chmod -R 777 /tmp/nginx /var/log/nginx /var/lib/nginx

EXPOSE 10000

# Inline startup: no external script, no shebang issues.
# Shell form CMD ensures /bin/sh -c wraps everything.
CMD set -e; \
    PORT="${PORT:-10000}"; \
    mkdir -p /tmp/nginx; \
    sed "s/PORT_PLACEHOLDER/${PORT}/g" /app/nginx.conf.template > /tmp/nginx/nginx.conf; \
    /app/posa-backend & \
    /app/ai/.venv/bin/python -m uvicorn ai.evaluator:app --host 127.0.0.1 --port 8000 --log-level warning & \
    cd /app/frontend && npx next start -p 3000 & \
    cd /app; \
    sleep 3; \
    echo "PoSA ready on port ${PORT}"; \
    exec nginx -c /tmp/nginx/nginx.conf -g "daemon off;"
