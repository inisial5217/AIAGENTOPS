# Builder stage
FROM python:3.12-slim AS builder

ENV PYTHONUNBUFFERED=1 \
    PYTHONDONTWRITEBYTECODE=1 \
    PIP_NO_CACHE_DIR=1

WORKDIR /install

RUN apt-get update && apt-get install -y --no-install-recommends \
    gcc \
    && rm -rf /var/lib/apt/lists/*

COPY requirements.txt .
RUN pip install --prefix=/install/dist --no-cache-dir -r requirements.txt

# Final runtime stage
FROM python:3.12-slim AS runner

ENV PYTHONUNBUFFERED=1 \
    PYTHONDONTWRITEBYTECODE=1 \
    PATH="/app/dist/bin:$PATH" \
    PYTHONPATH="/app/dist/lib/python3.12/site-packages:$PYTHONPATH"

WORKDIR /app

RUN groupadd -r appgroup && useradd -r -g appgroup -u 10001 appuser

COPY --from=builder /install/dist /app/dist
COPY app/ app/

USER appuser

EXPOSE 8000 50051

HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
    CMD python -c "import urllib.request; urllib.request.urlopen('http://localhost:8000/healthz')" || exit 1

CMD ["uvicorn", "app.main:app", "--host", "0.0.0.0", "--port", "8000"]
