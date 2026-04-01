# PoSA Dockerfile
#
# Multi-stage build producing a single container with:
#   - Go backend on :8080
#   - Python AI engine on :8000
#   - Next.js frontend on :3000
#   - nginx reverse proxy on $PORT (default :80)
#
# Build:  docker build -t posa .
# Run:    docker run -p 80:80 --env-file .env posa

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
    nodejs \
    npm \
    curl \
    supervisor \
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

# --- Config files ---
COPY nginx.conf /etc/nginx/nginx.conf.template
COPY supervisord.conf /etc/supervisor/conf.d/posa.conf
COPY entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh

EXPOSE 80

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD curl -f http://localhost:${PORT:-80}/health || exit 1

ENTRYPOINT ["/app/entrypoint.sh"]
