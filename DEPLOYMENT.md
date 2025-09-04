# Hubigr - Руководство по развертыванию

## 📋 Содержание

1. [Требования](#требования)
2. [Локальная разработка](#локальная-разработка)
3. [Staging окружение](#staging-окружение)
4. [Production развертывание](#production-развертывание)
5. [Мониторинг и логирование](#мониторинг-и-логирование)
6. [Резервное копирование](#резервное-копирование)
7. [Troubleshooting](#troubleshooting)

---

## 🔧 Требования

### Минимальные системные требования

**Разработка:**
- CPU: 2 ядра
- RAM: 4 GB
- Диск: 10 GB
- Docker 20.10+
- Docker Compose 2.0+

**Production:**
- CPU: 4 ядра
- RAM: 8 GB
- Диск: 50 GB SSD
- Docker Swarm или Kubernetes
- Load Balancer (nginx/HAProxy)

### Внешние зависимости

- **SMTP сервер** (Gmail, SendGrid, AWS SES)
- **Домен** с SSL сертификатом
- **Мониторинг** (опционально)

---

## 💻 Локальная разработка

### Быстрый старт

```bash
# Клонирование
git clone <repository-url>
cd hubigr

# Настройка
cp .env.example .env

# Запуск
docker-compose up -d

# Проверка
curl http://localhost:8000/api/v1/health
```

### Разработка без Docker

```bash
# Запуск только БД и Redis
docker-compose up -d postgres redis

# Установка зависимостей
go mod download

# Настройка переменных
export DATABASE_URL="postgres://user:pass@localhost:5432/hubigr?sslmode=disable"
export REDIS_URL="redis://localhost:6379"
export JWT_SECRET="dev-secret-key"

# Запуск сервиса
go run cmd/auth/main.go
```

### Горячая перезагрузка

```bash
# Установка air
go install github.com/cosmtrek/air@latest

# Запуск с автоперезагрузкой
air -c .air.toml
```

**Файл `.air.toml`:**
```toml
root = "."
testdata_dir = "testdata"
tmp_dir = "tmp"

[build]
  args_bin = []
  bin = "./tmp/main"
  cmd = "go build -o ./tmp/main ./cmd/auth"
  delay = 1000
  exclude_dir = ["assets", "tmp", "vendor", "testdata"]
  exclude_file = []
  exclude_regex = ["_test.go"]
  exclude_unchanged = false
  follow_symlink = false
  full_bin = ""
  include_dir = []
  include_ext = ["go", "tpl", "tmpl", "html"]
  kill_delay = "0s"
  log = "build-errors.log"
  send_interrupt = false
  stop_on_root = false

[color]
  app = ""
  build = "yellow"
  main = "magenta"
  runner = "green"
  watcher = "cyan"

[log]
  time = false

[misc]
  clean_on_exit = false
```

---

## 🧪 Staging окружение

### Docker Compose для Staging

**docker-compose.staging.yml:**
```yaml
version: '3.8'

services:
  postgres:
    image: postgres:15
    environment:
      POSTGRES_DB: hubigr_staging
      POSTGRES_USER: ${DB_USER}
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    volumes:
      - postgres_staging_data:/var/lib/postgresql/data
      - ./migrations:/docker-entrypoint-initdb.d
    restart: unless-stopped

  redis:
    image: redis:7-alpine
    command: redis-server --requirepass ${REDIS_PASSWORD}
    restart: unless-stopped

  auth:
    build:
      context: .
      dockerfile: Dockerfile.auth
    environment:
      PORT: 8080
      DATABASE_URL: postgres://${DB_USER}:${DB_PASSWORD}@postgres:5432/hubigr_staging?sslmode=disable
      REDIS_URL: redis://:${REDIS_PASSWORD}@redis:6379
      JWT_SECRET: ${JWT_SECRET}
      BASE_URL: https://staging.hubigr.com
      SMTP_HOST: ${SMTP_HOST}
      SMTP_USER: ${SMTP_USER}
      SMTP_PASS: ${SMTP_PASS}
      SMTP_FROM: noreply@staging.hubigr.com
    depends_on:
      - postgres
      - redis
    restart: unless-stopped

  krakend:
    image: devopsfaith/krakend:2.5
    ports:
      - "8000:8000"
    volumes:
      - ./krakend:/etc/krakend
    depends_on:
      - auth
    restart: unless-stopped

  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx/staging.conf:/etc/nginx/nginx.conf
      - ./ssl:/etc/ssl/certs
    depends_on:
      - krakend
    restart: unless-stopped

volumes:
  postgres_staging_data:
```

### Staging .env

```bash
# Database
DB_USER=hubigr_staging
DB_PASSWORD=secure_staging_password

# Redis
REDIS_PASSWORD=redis_staging_password

# JWT
JWT_SECRET=staging-jwt-secret-key-very-long

# SMTP
SMTP_HOST=smtp.gmail.com
SMTP_USER=staging@hubigr.com
SMTP_PASS=staging_smtp_password

# SSL
SSL_CERT_PATH=/etc/ssl/certs/staging.crt
SSL_KEY_PATH=/etc/ssl/certs/staging.key
```

### Запуск Staging

```bash
# Создание .env для staging
cp .env.staging .env

# Запуск
docker-compose -f docker-compose.yml -f docker-compose.staging.yml up -d

# Проверка
curl https://staging.hubigr.com/api/v1/health
```

---

## 🚀 Production развертывание

### Подготовка сервера

```bash
# Обновление системы
sudo apt update && sudo apt upgrade -y

# Установка Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo usermod -aG docker $USER

# Установка Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# Настройка firewall
sudo ufw allow 22
sudo ufw allow 80
sudo ufw allow 443
sudo ufw enable
```

### SSL сертификаты

```bash
# Установка Certbot
sudo apt install certbot

# Получение сертификата
sudo certbot certonly --standalone -d hubigr.com -d www.hubigr.com

# Автообновление
echo "0 12 * * * /usr/bin/certbot renew --quiet" | sudo crontab -
```

### Production Docker Compose

**docker-compose.prod.yml:**
```yaml
version: '3.8'

services:
  postgres:
    image: postgres:15
    environment:
      POSTGRES_DB: hubigr
      POSTGRES_USER: ${DB_USER}
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./migrations:/docker-entrypoint-initdb.d
      - ./backups:/backups
    restart: unless-stopped
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"

  redis:
    image: redis:7-alpine
    command: redis-server --requirepass ${REDIS_PASSWORD} --maxmemory 256mb --maxmemory-policy allkeys-lru
    volumes:
      - redis_data:/data
    restart: unless-stopped
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"

  auth:
    build:
      context: .
      dockerfile: Dockerfile.prod
    environment:
      PORT: 8080
      DATABASE_URL: postgres://${DB_USER}:${DB_PASSWORD}@postgres:5432/hubigr?sslmode=disable
      REDIS_URL: redis://:${REDIS_PASSWORD}@redis:6379
      JWT_SECRET: ${JWT_SECRET}
      BASE_URL: https://hubigr.com
      SMTP_HOST: ${SMTP_HOST}
      SMTP_USER: ${SMTP_USER}
      SMTP_PASS: ${SMTP_PASS}
      SMTP_FROM: noreply@hubigr.com
    volumes:
      - ./uploads:/app/uploads
    depends_on:
      - postgres
      - redis
    restart: unless-stopped
    deploy:
      replicas: 3
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"

  krakend:
    image: devopsfaith/krakend:2.5
    volumes:
      - ./krakend:/etc/krakend
    depends_on:
      - auth
    restart: unless-stopped
    deploy:
      replicas: 2

  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx/prod.conf:/etc/nginx/nginx.conf
      - /etc/letsencrypt:/etc/letsencrypt:ro
      - ./uploads:/var/www/uploads:ro
    depends_on:
      - krakend
    restart: unless-stopped

volumes:
  postgres_data:
  redis_data:
```

### Production Dockerfile

**Dockerfile.prod:**
```dockerfile
# Multi-stage build
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/auth

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

COPY --from=builder /app/main .
COPY --from=builder /app/uploads ./uploads

EXPOSE 8080

CMD ["./main"]
```

### Nginx конфигурация

**nginx/prod.conf:**
```nginx
events {
    worker_connections 1024;
}

http {
    upstream krakend {
        server krakend:8000;
    }

    # Rate limiting
    limit_req_zone $binary_remote_addr zone=api:10m rate=10r/s;
    limit_req_zone $binary_remote_addr zone=auth:10m rate=5r/s;

    # SSL redirect
    server {
        listen 80;
        server_name hubigr.com www.hubigr.com;
        return 301 https://$server_name$request_uri;
    }

    # Main server
    server {
        listen 443 ssl http2;
        server_name hubigr.com www.hubigr.com;

        # SSL
        ssl_certificate /etc/letsencrypt/live/hubigr.com/fullchain.pem;
        ssl_certificate_key /etc/letsencrypt/live/hubigr.com/privkey.pem;
        ssl_protocols TLSv1.2 TLSv1.3;
        ssl_ciphers ECDHE-RSA-AES256-GCM-SHA512:DHE-RSA-AES256-GCM-SHA512;
        ssl_prefer_server_ciphers off;

        # Security headers
        add_header X-Frame-Options DENY;
        add_header X-Content-Type-Options nosniff;
        add_header X-XSS-Protection "1; mode=block";
        add_header Strict-Transport-Security "max-age=31536000; includeSubDomains";

        # API proxy
        location /api/ {
            limit_req zone=api burst=20 nodelay;
            proxy_pass http://krakend;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
        }

        # Auth endpoints with stricter limits
        location /api/v1/auth/ {
            limit_req zone=auth burst=10 nodelay;
            proxy_pass http://krakend;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
        }

        # Static files
        location /uploads/ {
            alias /var/www/uploads/;
            expires 1y;
            add_header Cache-Control "public, immutable";
        }

        # Frontend (если есть)
        location / {
            root /var/www/html;
            try_files $uri $uri/ /index.html;
        }
    }
}
```

### Production .env

```bash
# Database
DB_USER=hubigr_prod
DB_PASSWORD=very_secure_production_password_123

# Redis
REDIS_PASSWORD=very_secure_redis_password_456

# JWT (минимум 32 символа)
JWT_SECRET=super-secure-jwt-secret-key-for-production-environment-2024

# SMTP
SMTP_HOST=smtp.sendgrid.net
SMTP_USER=apikey
SMTP_PASS=SG.your_sendgrid_api_key
```

### Запуск Production

```bash
# Создание директорий
sudo mkdir -p /opt/hubigr
cd /opt/hubigr

# Клонирование кода
git clone <repository-url> .

# Настройка окружения
sudo cp .env.production .env
sudo chown root:root .env
sudo chmod 600 .env

# Запуск
sudo docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d

# Проверка
curl https://hubigr.com/api/v1/health
```

---

## 📊 Мониторинг и логирование

### Prometheus + Grafana

**docker-compose.monitoring.yml:**
```yaml
version: '3.8'

services:
  prometheus:
    image: prom/prometheus
    ports:
      - "9090:9090"
    volumes:
      - ./monitoring/prometheus.yml:/etc/prometheus/prometheus.yml
      - prometheus_data:/prometheus
    restart: unless-stopped

  grafana:
    image: grafana/grafana
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin123
    volumes:
      - grafana_data:/var/lib/grafana
      - ./monitoring/grafana:/etc/grafana/provisioning
    restart: unless-stopped

  node-exporter:
    image: prom/node-exporter
    ports:
      - "9100:9100"
    restart: unless-stopped

volumes:
  prometheus_data:
  grafana_data:
```

### ELK Stack для логов

**docker-compose.logging.yml:**
```yaml
version: '3.8'

services:
  elasticsearch:
    image: docker.elastic.co/elasticsearch/elasticsearch:8.5.0
    environment:
      - discovery.type=single-node
      - "ES_JAVA_OPTS=-Xms512m -Xmx512m"
    volumes:
      - elasticsearch_data:/usr/share/elasticsearch/data
    ports:
      - "9200:9200"

  kibana:
    image: docker.elastic.co/kibana/kibana:8.5.0
    ports:
      - "5601:5601"
    depends_on:
      - elasticsearch

  logstash:
    image: docker.elastic.co/logstash/logstash:8.5.0
    volumes:
      - ./logging/logstash.conf:/usr/share/logstash/pipeline/logstash.conf
    depends_on:
      - elasticsearch

volumes:
  elasticsearch_data:
```

### Health checks

```bash
#!/bin/bash
# health-check.sh

# Проверка API
if ! curl -f http://localhost:8000/api/v1/health > /dev/null 2>&1; then
    echo "API health check failed"
    exit 1
fi

# Проверка БД
if ! docker-compose exec postgres pg_isready -U hubigr_prod > /dev/null 2>&1; then
    echo "Database health check failed"
    exit 1
fi

# Проверка Redis
if ! docker-compose exec redis redis-cli ping > /dev/null 2>&1; then
    echo "Redis health check failed"
    exit 1
fi

echo "All services healthy"
```

---

## 💾 Резервное копирование

### Автоматические бэкапы БД

```bash
#!/bin/bash
# backup.sh

DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_DIR="/opt/hubigr/backups"
DB_NAME="hubigr"

# Создание бэкапа
docker-compose exec postgres pg_dump -U hubigr_prod $DB_NAME | gzip > $BACKUP_DIR/db_backup_$DATE.sql.gz

# Удаление старых бэкапов (старше 30 дней)
find $BACKUP_DIR -name "db_backup_*.sql.gz" -mtime +30 -delete

# Загрузка в S3 (опционально)
# aws s3 cp $BACKUP_DIR/db_backup_$DATE.sql.gz s3://hubigr-backups/
```

### Cron задача для бэкапов

```bash
# Добавить в crontab
0 2 * * * /opt/hubigr/backup.sh >> /var/log/hubigr-backup.log 2>&1
```

### Восстановление из бэкапа

```bash
#!/bin/bash
# restore.sh

BACKUP_FILE=$1

if [ -z "$BACKUP_FILE" ]; then
    echo "Usage: $0 <backup_file>"
    exit 1
fi

# Остановка сервисов
docker-compose stop auth

# Восстановление БД
gunzip -c $BACKUP_FILE | docker-compose exec -T postgres psql -U hubigr_prod hubigr

# Запуск сервисов
docker-compose start auth
```

---

## 🔧 Troubleshooting

### Частые проблемы

**1. Сервис не запускается**
```bash
# Проверка логов
docker-compose logs auth

# Проверка ресурсов
docker stats

# Проверка портов
netstat -tulpn | grep :8080
```

**2. Проблемы с БД**
```bash
# Подключение к БД
docker-compose exec postgres psql -U hubigr_prod hubigr

# Проверка соединений
SELECT * FROM pg_stat_activity;

# Проверка размера БД
SELECT pg_size_pretty(pg_database_size('hubigr'));
```

**3. Проблемы с Redis**
```bash
# Подключение к Redis
docker-compose exec redis redis-cli

# Проверка памяти
INFO memory

# Очистка кеша
FLUSHALL
```

**4. SSL проблемы**
```bash
# Проверка сертификата
openssl x509 -in /etc/letsencrypt/live/hubigr.com/fullchain.pem -text -noout

# Обновление сертификата
sudo certbot renew

# Перезагрузка nginx
docker-compose restart nginx
```

### Мониторинг производительности

```bash
# Использование ресурсов
docker stats

# Логи в реальном времени
docker-compose logs -f auth

# Проверка дискового пространства
df -h

# Проверка сетевых соединений
ss -tulpn
```

### Скрипты для автоматизации

**deploy.sh:**
```bash
#!/bin/bash
set -e

echo "Starting deployment..."

# Обновление кода
git pull origin main

# Сборка новых образов
docker-compose build

# Обновление сервисов без даунтайма
docker-compose up -d --no-deps auth

# Проверка здоровья
sleep 10
if curl -f http://localhost:8000/api/v1/health; then
    echo "Deployment successful"
else
    echo "Deployment failed, rolling back..."
    docker-compose restart auth
    exit 1
fi
```

---

## 📞 Поддержка

### Контакты для экстренных случаев

- **DevOps**: [email]
- **Backend**: [email]
- **Мониторинг**: [Grafana URL]
- **Логи**: [Kibana URL]

### Полезные команды

```bash
# Быстрая диагностика
docker-compose ps
docker-compose logs --tail=50 auth
curl -I http://localhost:8000/api/v1/health

# Перезапуск сервисов
docker-compose restart auth
docker-compose restart nginx

# Обновление без даунтайма
docker-compose up -d --no-deps auth
```

---

*Deployment Guide v1.0 - Обновлено: $(date)*