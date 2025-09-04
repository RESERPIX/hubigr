# Hubigr - Платформа для геймджемов

Веб-платформа для организации и проведения игровых джемов согласно техническому заданию.

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
- ⏳ 1.3 Игры и запуск в браузере
- ⏳ 1.4 Модуль «Геймджем»
- ⏳ 1.5 Коммуникации и уведомления
- ⏳ 1.6 Админ-минимум

## API Endpoints

### Аутентификация
- `POST /api/v1/auth/signup` - Регистрация (UC-1.1.1)
- `POST /api/v1/auth/login` - Вход (UC-1.1.2)
- `POST /api/v1/auth/verify-email` - Подтверждение email (UC-1.1.1)
- `POST /api/v1/auth/resend-verification` - Повторная отправка письма
- `POST /api/v1/auth/reset-password` - Запрос сброса пароля (UC-1.1.3)
- `POST /api/v1/auth/reset-password/confirm` - Подтверждение сброса (UC-1.1.3)
- `POST /api/v1/auth/logout` - Выход (UC-1.1.4)

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

# Тест rate limiting (5 попыток/мин)
./test_ratelimit.sh

# Тест email verification
./test_email.sh

# Тест восстановления пароля
./test_reset_password.sh

# Тест списка сабмитов
./test_submissions.sh

# Тест загрузки аватара
./test_avatar.sh
```

## Валидация согласно ТЗ

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
- **Rate limiting с Redis: 5 попыток входа/мин** ✅
- **Email отправка для verification (SMTP + Mock)** ✅
- **Восстановление пароля (UC-1.1.3)** ✅
- **Список сабмитов пользователя (UC-1.2.2)** ✅
- **Загрузка аватаров через API** ✅
- Валидация всех входных данных
- CORS настроен

## База данных

Модели согласно ТЗ:
- `users` (DM-3.1)
- `notification_settings` (DM-3.15)
- `follows` (DM-3.14)
- `email_verify_tokens`
- `password_reset_tokens`
- `user_submissions` (заглушка для UC-1.2.2)