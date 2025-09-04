# Hubigr API Documentation

## Базовая информация

- **Base URL**: `http://localhost:8000/api/v1`
- **Формат**: JSON
- **Авторизация**: Bearer JWT Token
- **Rate Limiting**: 5 запросов/минуту для auth endpoints

---

## 🔐 Аутентификация

### Регистрация

**POST** `/auth/signup`

```json
{
  "email": "user@example.com",
  "password": "password123",
  "confirm_password": "password123",
  "nick": "username",
  "agree_terms": true
}
```

**Ответ 201:**
```json
{
  "message": "Мы отправили ссылку для подтверждения на email",
  "user_id": 1
}
```

**Ошибки:**
- `409` - Email уже зарегистрирован
- `422` - Ошибка валидации
- `429` - Превышен лимит запросов

---

### Вход в систему

**POST** `/auth/login`

```json
{
  "email": "user@example.com", 
  "password": "password123"
}
```

**Ответ 200:**
```json
{
  "user": {
    "id": 1,
    "email": "user@example.com",
    "nick": "username",
    "role": "participant",
    "avatar": "https://example.com/avatar.jpg",
    "bio": "Описание пользователя",
    "email_verified": true,
    "created_at": "2024-01-15T10:30:00Z"
  },
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Ошибки:**
- `401` - Неверные учетные данные
- `401` - Email не подтвержден
- `403` - Аккаунт заблокирован

---

### Подтверждение email

**POST** `/auth/verify-email?token=<verification_token>`

**Ответ 200:**
```json
{
  "message": "Email успешно подтвержден"
}
```

---

### Повторная отправка подтверждения

**POST** `/auth/resend-verification`

```json
{
  "email": "user@example.com"
}
```

**Ответ 200:**
```json
{
  "message": "Новая ссылка отправлена на email"
}
```

---

### Сброс пароля

**POST** `/auth/reset-password`

```json
{
  "email": "user@example.com"
}
```

**Ответ 200:**
```json
{
  "message": "Мы отправили ссылку на email"
}
```

---

### Подтверждение сброса пароля

**POST** `/auth/reset-password/confirm`

```json
{
  "token": "reset_token_here",
  "password": "newpassword123",
  "confirm_password": "newpassword123"
}
```

**Ответ 200:**
```json
{
  "message": "Пароль успешно изменен"
}
```

---

### Выход из системы

**POST** `/auth/logout`

**Headers:** `Authorization: Bearer <token>`

**Ответ 200:**
```json
{
  "message": "Вы успешно вышли из системы"
}
```

---

## 👤 Профили

> Все endpoints профилей требуют авторизации

### Получение профиля

**GET** `/profile`

**Headers:** `Authorization: Bearer <token>`

**Ответ 200:**
```json
{
  "id": 1,
  "email": "user@example.com",
  "nick": "username",
  "role": "participant",
  "avatar": "https://example.com/avatar.jpg",
  "bio": "Описание пользователя",
  "links": [
    {
      "title": "GitHub",
      "url": "https://github.com/username"
    }
  ],
  "email_verified": true,
  "created_at": "2024-01-15T10:30:00Z",
  "privacy_settings": {
    "show_email": false,
    "show_submissions": true
  }
}
```

---

### Обновление профиля

**PUT** `/profile`

**Headers:** `Authorization: Bearer <token>`

```json
{
  "nick": "newnickname",
  "bio": "Новое описание профиля",
  "links": [
    {
      "title": "GitHub", 
      "url": "https://github.com/newusername"
    },
    {
      "title": "Twitter",
      "url": "https://twitter.com/newusername"
    }
  ],
  "privacy_settings": {
    "show_email": false,
    "show_submissions": true
  }
}
```

**Ответ 200:**
```json
{
  "message": "Профиль обновлен"
}
```

**Ошибки:**
- `422` - Ошибка валидации (ник 2-50 символов, био до 200 символов, максимум 5 ссылок)

---

### Загрузка аватара

**POST** `/profile/avatar`

**Headers:** 
- `Authorization: Bearer <token>`
- `Content-Type: multipart/form-data`

**Body:** `avatar: <image_file>`

**Ответ 200:**
```json
{
  "avatar_url": "http://localhost:3000/uploads/avatars/1_abc123def456.jpg",
  "message": "Аватар обновлен"
}
```

**Ограничения:**
- Форматы: JPEG, PNG
- Максимальный размер: 2 МБ

**Ошибки:**
- `400` - Файл не найден
- `422` - Неподдерживаемый формат или размер

---

### Настройки уведомлений

**GET** `/profile/notifications`

**Headers:** `Authorization: Bearer <token>`

**Ответ 200:**
```json
{
  "user_id": 1,
  "new_game": true,
  "new_build": true, 
  "new_post": false,
  "channels": {
    "in_app": true,
    "email": false,
    "push": true
  }
}
```

---

**PUT** `/profile/notifications`

**Headers:** `Authorization: Bearer <token>`

```json
{
  "new_game": true,
  "new_build": false,
  "new_post": true,
  "channels": {
    "in_app": true,
    "email": true,
    "push": false
  }
}
```

**Ответ 200:**
```json
{
  "message": "Настройки обновлены"
}
```

---

### Список сабмитов пользователя

**GET** `/profile/submissions?page=1&limit=20`

**Headers:** `Authorization: Bearer <token>`

**Query Parameters:**
- `page` (optional): Номер страницы (по умолчанию 1)
- `limit` (optional): Количество на странице (по умолчанию 20, максимум 100)

**Ответ 200:**
```json
{
  "submissions": [
    {
      "id": 1,
      "jam_title": "Spring Game Jam 2024",
      "jam_slug": "spring-jam-2024",
      "game_title": "My Awesome Game", 
      "game_slug": "my-awesome-game",
      "status": "active",
      "submitted_at": "2024-01-15T10:30:00Z"
    }
  ],
  "total": 5,
  "page": 1,
  "limit": 20
}
```

---

## 🔧 Служебные endpoints

### Health Check

**GET** `/health`

**Ответ 200:**
```json
{
  "status": "ok",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

---

## 📁 Статические файлы

### Доступ к загруженным файлам

**GET** `/uploads/avatars/{filename}`

Прямой доступ к загруженным аватарам.

**Пример:**
```
GET /uploads/avatars/1_abc123def456.jpg
```

---

## ❌ Коды ошибок

### Стандартный формат ошибки

```json
{
  "error": {
    "code": "error_code",
    "message": "Описание ошибки"
  }
}
```

### Коды ошибок

| Код | Описание | Примеры |
|-----|----------|---------|
| `400` | Bad Request | Неверный формат JSON |
| `401` | Unauthorized | Неверный токен, email не подтвержден |
| `403` | Forbidden | Аккаунт заблокирован |
| `404` | Not Found | Пользователь не найден |
| `409` | Conflict | Email уже зарегистрирован |
| `422` | Validation Error | Неверные данные формы |
| `429` | Too Many Requests | Превышен лимит запросов |
| `500` | Internal Server Error | Внутренняя ошибка сервера |

### Специфичные коды ошибок

- `email_not_verified` - Email не подтвержден
- `already_verified` - Email уже подтвержден  
- `invalid_token` - Недействительный токен
- `upload_error` - Ошибка загрузки файла
- `validation_error` - Ошибка валидации данных

---

## 🔒 Авторизация

### JWT Token

Токен передается в заголовке:
```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

### Структура токена

```json
{
  "user_id": 1,
  "role": "participant", 
  "nick": "username",
  "exp": 1642248000,
  "iat": 1642161600
}
```

### Время жизни

- **Access Token**: 24 часа
- **Email Verification**: 1 час
- **Password Reset**: 1 час

---

## 📊 Rate Limiting

### Лимиты

| Endpoint | Лимит | Период |
|----------|-------|--------|
| `/auth/login` | 5 запросов | 1 минута |
| `/auth/signup` | 5 запросов | 1 минута |
| `/auth/reset-password` | 5 запросов | 1 минута |
| `/auth/resend-verification` | 5 запросов | 1 минута |

### Заголовки ответа

```
X-RateLimit-Limit: 5
X-RateLimit-Remaining: 3
X-RateLimit-Reset: 1642161660
```

---

## 🧪 Примеры использования

### Полный цикл регистрации

```bash
# 1. Регистрация
curl -X POST http://localhost:8000/api/v1/auth/signup \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "testpass123", 
    "confirm_password": "testpass123",
    "nick": "testuser",
    "agree_terms": true
  }'

# 2. Подтверждение email (токен из письма)
curl -X POST "http://localhost:8000/api/v1/auth/verify-email?token=<token>"

# 3. Вход в систему
curl -X POST http://localhost:8000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "testpass123"
  }'

# 4. Получение профиля
curl -X GET http://localhost:8000/api/v1/profile \
  -H "Authorization: Bearer <jwt_token>"
```

### Загрузка аватара

```bash
curl -X POST http://localhost:8000/api/v1/profile/avatar \
  -H "Authorization: Bearer <jwt_token>" \
  -F "avatar=@/path/to/image.jpg"
```

---

*API Documentation v1.0 - Обновлено: $(date)*