# Hubigr - Документация проекта

## 📋 Содержание

1. [Обзор проекта](#обзор-проекта)
2. [Архитектура](#архитектура)
3. [Установка и запуск](#установка-и-запуск)
4. [API Reference](#api-reference)
5. [База данных](#база-данных)
6. [Безопасность](#безопасность)
7. [Конфигурация](#конфигурация)
8. [Тестирование](#тестирование)
9. [Развертывание](#развертывание)
10. [Разработка](#разработка)

---

## 🎯 Обзор проекта

**Hubigr** - веб-платформа для организации и проведения игровых джемов (game jams). Проект реализован как микросервисная архитектура с API Gateway.

### Статус реализации

- ✅ **1.1 Аккаунты и доступ** - полностью реализован
- ✅ **1.2 Профили** - полностью реализован  
- ⏳ **1.3 Игры и запуск в браузере** - планируется
- ⏳ **1.4 Модуль «Геймджем»** - планируется
- ⏳ **1.5 Коммуникации и уведомления** - планируется
- ⏳ **1.6 Админ-минимум** - планируется

### Технологический стек

- **Backend**: Go 1.21+
- **API Gateway**: KrakenD
- **База данных**: PostgreSQL 15
- **Кеширование**: Redis 7
- **Контейнеризация**: Docker & Docker Compose
- **Web Framework**: Fiber v2

---

## 🏗️ Архитектура

### Компоненты системы

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Frontend      │───▶│   KrakenD       │───▶│   Auth Service  │
│   (React/Vue)   │    │   API Gateway   │    │   (Go + Fiber)  │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                │                       │
                                │                       ▼
                                │              ┌─────────────────┐
                                │              │   PostgreSQL    │
                                │              │   Database      │
                                │              └─────────────────┘
                                │                       │
                                ▼                       ▼
                       ┌─────────────────┐    ┌─────────────────┐
                       │     Redis       │    │   File Storage  │
                       │   (Rate Limit)  │    │   (Avatars)     │
                       └─────────────────┘    └─────────────────┘
```

### Структура проекта

```
hubigr/
├── cmd/
│   └── auth/                 # Точка входа auth сервиса
├── internal/
│   ├── config/              # Конфигурация
│   ├── domain/              # Доменные модели
│   ├── email/               # Email сервис (SMTP/Mock)
│   ├── http/                # HTTP handlers и middleware
│   ├── ratelimit/           # Rate limiting с Redis
│   ├── security/            # JWT, bcrypt, токены
│   ├── store/               # Репозитории БД
│   ├── upload/              # Загрузка файлов
│   └── validation/          # Валидация данных
├── krakend/                 # Конфигурация API Gateway
├── migrations/              # SQL миграции
├── uploads/                 # Загруженные файлы
├── docker-compose.yml       # Оркестрация сервисов
├── Dockerfile.auth          # Образ auth сервиса
└── *.sh                     # Тестовые скрипты
```

---

## 🚀 Установка и запуск

### Требования

- Docker 20.10+
- Docker Compose 2.0+
- Go 1.21+ (для разработки)

### Быстрый старт

```bash
# 1. Клонирование репозитория
git clone <repository-url>
cd hubigr

# 2. Настройка конфигурации
cp .env.example .env
# Отредактируйте .env для SMTP настроек (опционально)

# 3. Запуск всех сервисов
docker-compose up -d

# 4. Проверка статуса
docker-compose ps

# 5. Проверка API
curl http://localhost:8000/api/v1/health
```

### Порты сервисов

- **8000** - KrakenD API Gateway (основной вход)
- **8080** - Auth Service (внутренний)
- **5432** - PostgreSQL
- **6379** - Redis

---

## 📡 API Reference

### Базовый URL
```
http://localhost:8000/api/v1
```

### Аутентификация

#### Регистрация
```http
POST /auth/signup
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "<your_password>",
  "confirm_password": "<your_password>", 
  "nick": "username",
  "agree_terms": true
}
```

**Ответ:**
```json
{
  "message": "Мы отправили ссылку для подтверждения на email",
  "user_id": 1
}
```

#### Вход в систему
```http
POST /auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "<your_password>"
}
```

**Ответ:**
```json
{
  "user": {
    "id": 1,
    "email": "user@example.com",
    "nick": "username",
    "role": "participant",
    "email_verified": true
  },
  "token": "eyJhbGciOiJIUzI1NiIs..."
}
```

#### Подтверждение email
```http
POST /auth/verify-email?token=<verification_token>
```

#### Сброс пароля
```http
POST /auth/reset-password
Content-Type: application/json

{
  "email": "user@example.com"
}
```

```http
POST /auth/reset-password/confirm
Content-Type: application/json

{
  "token": "<reset_token>",
  "password": "<new_password>",
  "confirm_password": "<new_password>"
}
```

### Профили (требуется авторизация)

#### Получение профиля
```http
GET /profile
Authorization: Bearer <jwt_token>
```

#### Обновление профиля
```http
PUT /profile
Authorization: Bearer <jwt_token>
Content-Type: application/json

{
  "nick": "newnick",
  "bio": "Описание профиля",
  "links": [
    {"title": "GitHub", "url": "https://github.com/user"}
  ]
}
```

#### Загрузка аватара
```http
POST /profile/avatar
Authorization: Bearer <jwt_token>
Content-Type: multipart/form-data

avatar: <image_file>
```

#### Список сабмитов
```http
GET /profile/submissions?page=1&limit=20
Authorization: Bearer <jwt_token>
```

### Коды ошибок

- **400** - Неверный запрос
- **401** - Не авторизован
- **403** - Доступ запрещен
- **404** - Не найдено
- **409** - Конфликт (дублирование)
- **422** - Ошибка валидации
- **429** - Превышен лимит запросов
- **500** - Внутренняя ошибка сервера

---

## 🗄️ База данных

### Схема БД

#### Таблица `users`
```sql
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    email CITEXT UNIQUE NOT NULL,
    hash TEXT NOT NULL,
    role TEXT NOT NULL DEFAULT 'participant',
    nick TEXT NOT NULL,
    avatar TEXT,
    bio TEXT,
    links JSONB NOT NULL DEFAULT '[]',
    is_banned BOOLEAN NOT NULL DEFAULT false,
    email_verified BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    privacy_settings JSONB NOT NULL DEFAULT '{}'
);
```

#### Таблица `notification_settings`
```sql
CREATE TABLE notification_settings (
    user_id BIGINT PRIMARY KEY REFERENCES users(id),
    new_game BOOLEAN NOT NULL DEFAULT true,
    new_build BOOLEAN NOT NULL DEFAULT true,
    new_post BOOLEAN NOT NULL DEFAULT true,
    channels JSONB NOT NULL DEFAULT '{"in_app": true}'
);
```

#### Токены
- `email_verify_tokens` - токены подтверждения email (TTL 1 час)
- `password_reset_tokens` - токены сброса пароля (TTL 1 час)

### Миграции

Миграции выполняются автоматически при запуске PostgreSQL контейнера из папки `/migrations/`.

---

## 🔒 Безопасность

### Аутентификация
- **JWT токены** с алгоритмом HS256
- **Время жизни**: 24 часа
- **Секрет**: настраивается через переменную окружения

### Пароли
- **Хеширование**: bcrypt с cost 12
- **Требования**: 6-20 символов, буквы, цифры, спецсимволы
- **Валидация**: регулярные выражения

### Rate Limiting
- **5 попыток входа/минуту** на IP
- **Реализация**: Redis с TTL
- **Endpoints**: login, signup, reset-password

### Валидация файлов
- **Аватары**: только JPEG/PNG, до 2 МБ
- **Проверка**: Content-Type и размер файла

### CORS
```yaml
allow_origins: ["*"]
allow_methods: ["GET", "POST", "PUT", "DELETE", "OPTIONS"]
allow_headers: ["Origin", "Authorization", "Content-Type"]
```

---

## ⚙️ Конфигурация

### Переменные окружения

```bash
# Основные настройки
PORT=8080
DATABASE_URL=postgres://user:pass@localhost:5432/hubigr?sslmode=disable
REDIS_URL=redis://localhost:6379
JWT_SECRET=your-super-secret-jwt-key

# Email настройки (SMTP)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your-email@gmail.com
SMTP_PASS=your-app-password
SMTP_FROM=noreply@hubigr.com

# Базовый URL для ссылок
BASE_URL=https://hubigr.com
```

### Конфигурация для разработки

```bash
# Копируем пример
cp .env.example .env

# Для разработки можно оставить SMTP пустыми
# Будет использоваться Mock sender
```

### Конфигурация для продакшена

```bash
# Обязательно установить
JWT_SECRET=very-long-random-secret-key
SMTP_HOST=your-smtp-server
SMTP_USER=your-smtp-user
SMTP_PASS=your-smtp-password
BASE_URL=https://your-domain.com
```

---

## 🧪 Тестирование

### Автоматические тесты

```bash
# Rate limiting
./test_ratelimit.sh

# Email verification
./test_email.sh

# Восстановление пароля
./test_reset_password.sh

# Список сабмитов
./test_submissions.sh

# Загрузка аватара
./test_avatar.sh
```

### Ручное тестирование

#### Регистрация пользователя
```bash
curl -X POST http://localhost:8000/api/v1/auth/signup \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "<your_password>",
    "confirm_password": "<your_password>",
    "nick": "testuser",
    "agree_terms": true
  }'
```

#### Проверка health endpoint
```bash
curl http://localhost:8000/api/v1/health
```

### Мониторинг логов

```bash
# Все сервисы
docker-compose logs -f

# Конкретный сервис
docker-compose logs -f auth
docker-compose logs -f postgres
docker-compose logs -f redis
```

---

## 🚢 Развертывание

### Docker Production

```bash
# Сборка образов
docker-compose build

# Запуск в продакшене
docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d
```

### Kubernetes (пример)

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: hubigr-auth
spec:
  replicas: 3
  selector:
    matchLabels:
      app: hubigr-auth
  template:
    metadata:
      labels:
        app: hubigr-auth
    spec:
      containers:
      - name: auth
        image: hubigr/auth:latest
        ports:
        - containerPort: 8080
        env:
        - name: JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: hubigr-secrets
              key: jwt-secret
```

### Мониторинг

Рекомендуется добавить:
- **Prometheus** для метрик
- **Grafana** для дашбордов
- **Jaeger** для трейсинга
- **ELK Stack** для логов

---

## 👨‍💻 Разработка

### Локальная разработка

```bash
# Установка зависимостей
go mod download

# Запуск только БД и Redis
docker-compose up -d postgres redis

# Запуск auth сервиса локально
go run cmd/auth/main.go
```

### Добавление новых endpoints

1. **Добавить handler** в `internal/http/handlers.go`
2. **Добавить маршрут** в `internal/http/router.go`
3. **Обновить KrakenD** в `krakend/krakend.json`
4. **Добавить тесты**

### Миграции БД

```bash
# Создать новую миграцию
touch migrations/002_new_feature.sql

# Добавить в docker-compose для автоприменения
```

### Структура кода

```go
// Пример нового handler
func (h *Handlers) NewEndpoint(c *fiber.Ctx) error {
    // 1. Валидация входных данных
    var req SomeRequest
    if err := c.BodyParser(&req); err != nil {
        return c.Status(400).JSON(domain.NewError("bad_request", "Неверный формат"))
    }
    
    // 2. Бизнес-логика
    result, err := h.someRepo.DoSomething(c.Context(), req)
    if err != nil {
        return c.Status(500).JSON(domain.NewError("internal_error", "Ошибка"))
    }
    
    // 3. Возврат результата
    return c.JSON(result)
}
```

### Соглашения

- **Языки**: Go для backend, комментарии на русском
- **Форматирование**: `gofmt`, `goimports`
- **Линтинг**: `golangci-lint`
- **Тесты**: `*_test.go` файлы
- **Коммиты**: Conventional Commits

---

## 📞 Поддержка

### Частые проблемы

**Q: Ошибка подключения к БД**
```
A: Проверьте, что PostgreSQL запущен:
docker-compose ps postgres
```

**Q: Rate limiting не работает**
```
A: Проверьте подключение к Redis:
docker-compose logs redis
```

**Q: Email не отправляются**
```
A: В разработке используется Mock sender.
Проверьте логи: docker-compose logs auth
```

### Логи и отладка

```bash
# Подробные логи auth сервиса
docker-compose logs -f auth

# Подключение к БД для отладки
docker-compose exec postgres psql -U user -d hubigr

# Подключение к Redis
docker-compose exec redis redis-cli
```

### Контакты

- **Разработчик**: [Ваше имя]
- **Email**: [ваш email]
- **Репозиторий**: [ссылка на репозиторий]

---

*Документация обновлена: $(date)*