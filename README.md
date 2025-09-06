## Архитектура

- **Go** - Backend сервисы
- **KrakenD** - API Gateway
- **PostgreSQL** - База данных
- **Redis** - Кеширование и rate limiting
- **Docker** - Контейнеризация

## Текущий статус (MVP)

Реализованы модули:
- ✅ 1.1 Аккаунты и доступ
- ✅ 1.2 Профили
- ✅ Frontend UI/UX
- ⏳ 1.3 Игры и запуск в браузере
- ⏳ 1.4 Модуль «Геймджем»
- ⏳ 1.5 Коммуникации и уведомления
- ⏳ 1.6 Админ-минимум

## API Endpoints

### Аутентификация
- `POST /api/v1/auth/signup` - Регистрация (UC-1.1.1)
- `POST /api/v1/auth/login` - Вход (UC-1.1.2) + refresh token
- `POST /api/v1/auth/refresh` - Обновление токенов с ротацией
- `POST /api/v1/auth/verify-email` - Подтверждение email (UC-1.1.1)
- `POST /api/v1/auth/resend-verification` - Повторная отправка письма
- `POST /api/v1/auth/reset-password` - Запрос сброса пароля (UC-1.1.3)
- `POST /api/v1/auth/reset-password/confirm` - Подтверждение сброса (UC-1.1.3)
- `POST /api/v1/auth/logout` - Выход + отзыв токенов (UC-1.1.4)

### Профили
- `GET /api/v1/profile` - Получить профиль (UC-1.2.2)
- `PUT /api/v1/profile` - Обновить профиль (UC-1.2.1)
- `GET /api/v1/profile/notifications` - Настройки уведомлений (UC-1.2.3)
- `PUT /api/v1/profile/notifications` - Обновить настройки (UC-1.2.3)
- `GET /api/v1/profile/submissions` - Список сабмитов пользователя (UC-1.2.2)
- `POST /api/v1/profile/avatar` - Загрузка аватара (UC-1.2.1)

## Запуск

```bash
# Копирование примера конфигурации
cp .env.example .env
# Отредактируйте .env для SMTP настроек

# Запуск всех сервисов
docker-compose up -d

# Проверка здоровья
curl http://localhost:8000/api/v1/health
```


### Регистрация (UC-1.1.1)
- Email: формат example@example.com
- Пароль: 6-20 символов, 0-9, A-Z, a-z, спецсимволы
- Ник: 2-50 символов, A-Z, a-z, А-Я, а-я
- Подтверждение email: TTL 1 час

### Профиль (UC-1.2.1)
- Ник: 2-50 символов
- Аватар: jpeg/png, до 2 МБ
- Био: до 200 символов
- Ссылки: до 5, формат URL

## Безопасность

- JWT токены с HS256
- Bcrypt для паролей
- **Refresh токены с ротацией** ✅
- **TTL политики: access 5-15 мин, refresh 7-30 дней** ✅
- **Rate limiting с Redis: 5 попыток входа/мин** ✅
- **Turnstile капча на регистрации** ✅
- **Email отправка для verification (SMTP)** ✅
- **Восстановление пароля (UC-1.1.3)** ✅
- **Список сабмитов пользователя (UC-1.2.2)** ✅
- **Загрузка аватаров через API** ✅
- **Безопасная раздача статических файлов** ✅
- Валидация всех входных данных
- CORS настроен

## База данных

Модели согласно ТЗ:
- `users` (DM-3.1)
- `notification_settings` (DM-3.15)
- `follows` (DM-3.14)
- `email_verify_tokens`
- `password_reset_tokens`
- `refresh_tokens`
- `user_submissions` (заглушка для UC-1.2.2)

## SMTP настройка

```bash
# Добавьте в .env для реальной отправки email:
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=ваш-email@gmail.com
SMTP_PASS=app-password-от-google
SMTP_FROM=noreply@hubigr.com

# Для Gmail нужен App Password:
# https://myaccount.google.com/apppasswords
```