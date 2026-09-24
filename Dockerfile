FROM python:3.12-slim

WORKDIR /app

ENV PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1 \
    SYSTEM_MONITOR_ROOT=/app \
    SYSTEM_MONITOR_CONFIG_DIR=/app/config \
    SYSTEM_MONITOR_DATA_DIR=/app/data

COPY requirements.txt pyproject.toml README.md ./
COPY system_monitor/ system_monitor/
COPY web/ web/
COPY config/ config/

RUN pip install --no-cache-dir -r requirements.txt \
    && pip install --no-cache-dir ".[hub]"

RUN mkdir -p data

EXPOSE 8080

CMD ["python", "-m", "system_monitor", "--mode", "hub", "--host", "0.0.0.0", "--port", "8080", "--proxy-headers"]
