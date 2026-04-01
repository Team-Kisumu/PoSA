# PoSA Dockerfile
#
# Multi-stage build producing a single container with:
#   - Go backend on :8080
#   - Python AI engine on :8000
#   - Next.js frontend on :3000
#   - nginx reverse proxy on $PORT (default :10000)
#
# Build:  docker build -t posa .
# Run:    docker run -p 8080:10000 --env-file .env posa

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
FROM debian:bookworm-slim AS runtime

RUN apt-get update && apt-get install -y --no-install-recommends \
    nginx \
    python3 \
    python3-pip \
    python3-venv \
    curl \
    ca-certificates \
    gnupg \
    && rm -rf /var/lib/apt/lists/*

# Install Node.js 22 from NodeSource (Next.js 16 requires Node >= 20.9.0).
RUN curl -fsSL https://deb.nodesource.com/setup_22.x | bash - \
    && apt-get install -y --no-install-recommends nodejs \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# --- Go backend binary ---
COPY --from=backend-build /posa-backend /app/backend/posa-backend

# --- Python AI engine ---
COPY ai/ /app/ai/
RUN python3 -m venv /app/ai/.venv \
    && /app/ai/.venv/bin/pip install --no-cache-dir -r /app/ai/requirements.txt

# --- Next.js frontend ---
COPY --from=frontend-build /build/.next /app/frontend/.next
COPY --from=frontend-build /build/public /app/frontend/public
COPY --from=frontend-build /build/package.json /app/frontend/package.json
COPY --from=frontend-build /build/node_modules /app/frontend/node_modules

# --- Config and entrypoint ---
COPY nginx.conf /app/nginx.conf.template
COPY entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh

# Make all runtime paths writable for non-root execution (Render requirement).
RUN mkdir -p /tmp/nginx /app/logs \
    && chmod -R 777 /tmp/nginx /app/logs /var/log/nginx /var/lib/nginx \
    && chmod -R 755 /app/backend /app/ai /app/frontend

EXPOSE 10000

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD curl -f http://localhost:${PORT:-10000}/health || exit 1

ENTRYPOINT ["/app/entrypoint.sh"]
