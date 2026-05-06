# PoSA Dockerfile
#
# Single-process architecture for Render:
# Go backend listens on $PORT, proxies /ai/* to Python and /* to Next.js.
#
# Build:  docker build -t posa .
# Run:    docker run -p 10000:10000 -e PORT=10000 -e RENDER=true posa

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
# Stage 2: Build Next.js frontend
# ============================================================
FROM node:22-bookworm-slim AS frontend-build

WORKDIR /build
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci --ignore-scripts
COPY frontend/ ./

ENV NEXT_PUBLIC_API_URL=""
ENV NEXT_PUBLIC_AI_URL="/ai"

RUN npm run build

# ============================================================
# Stage 3: Production runtime
# ============================================================
FROM node:22-bookworm-slim AS runtime

RUN apt-get update && apt-get install -y --no-install-recommends \
    python3 \
    python3-venv \
    curl \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app
ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1
ENV PYTHONPATH=/app
ENV RENDER=true

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

# Create start script inline.
RUN printf '#!/bin/sh\n\
export PYTHONPATH=/app\n\
export RENDER=true\n\
echo "Starting AI engine..."\n\
/app/ai/.venv/bin/python -m uvicorn ai.evaluator:app --host 127.0.0.1 --port 8000 --log-level warning &\n\
echo "Starting frontend..."\n\
cd /app/frontend && node node_modules/next/dist/bin/next start -p 3000 &\n\
cd /app\n\
echo "Starting backend on port ${PORT:-10000}..."\n\
exec /app/posa-backend\n' > /app/start.sh && chmod +x /app/start.sh

EXPOSE 10000

CMD ["/app/start.sh"]
