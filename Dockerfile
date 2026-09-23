FROM python:3.12-slim

WORKDIR /app

ENV PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1

COPY requirements.txt pyproject.toml README.md ./
COPY system_monitor/ system_monitor/
COPY web/ web/
COPY config/ config/

RUN pip install --no-cache-dir -r requirements.txt \
    && pip install --no-cache-dir .

RUN mkdir -p data

EXPOSE 8080

CMD ["python", "-m", "system_monitor", "--host", "0.0.0.0", "--port", "8080"]
