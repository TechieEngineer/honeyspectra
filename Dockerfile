FROM golang:1.23-alpine AS go-build
WORKDIR /src
COPY go.mod ./
COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 go build -trimpath -o /out/honeypot ./cmd/honeypot
FROM alpine:3.22 AS engine
RUN addgroup -g 10001 lab && adduser -D -u 10001 -G lab lab && mkdir -p /app/data && chown -R lab:lab /app
WORKDIR /app
COPY --from=go-build /out/honeypot /usr/local/bin/honeypot
COPY configs ./configs
USER 10001:10001
ENTRYPOINT ["honeypot"]
FROM python:3.13-slim AS dashboard
ENV PYTHONDONTWRITEBYTECODE=1 PYTHONUNBUFFERED=1
WORKDIR /app
COPY requirements.txt ./
RUN pip install --no-cache-dir -r requirements.txt && groupadd -g 10001 lab && useradd -u 10001 -g lab lab && mkdir -p /app/data && chown -R lab:lab /app
COPY backend ./backend
COPY simulator.py ./
USER 10001:10001
CMD ["python", "-m", "uvicorn", "backend.app:app", "--host", "0.0.0.0", "--port", "8000", "--limit-concurrency", "32"]
