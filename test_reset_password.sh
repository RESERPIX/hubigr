#!/bin/bash

# Тест восстановления пароля согласно ТЗ UC-1.1.3

echo "Тестирование восстановления пароля..."

# 1. Запрос сброса пароля для существующего пользователя
echo "1. Запрос сброса пароля:"
curl -X POST http://localhost:8000/api/v1/auth/reset-password \
    -H "Content-Type: application/json" \
    -d '{"email":"test@example.com"}' \
    -w "\nHTTP Status: %{http_code}\n" \
    -s | jq .
echo "---"

# 2. Запрос сброса для несуществующего email (не должен раскрывать существование)
echo "2. Запрос сброса для несуществующего email:"
curl -X POST http://localhost:8000/api/v1/auth/reset-password \
    -H "Content-Type: application/json" \
    -d '{"email":"nonexistent@example.com"}' \
    -w "\nHTTP Status: %{http_code}\n" \
    -s | jq .
echo "---"

# 3. Попытка подтверждения с неверным токеном
echo "3. Подтверждение с неверным токеном:"
curl -X POST http://localhost:8000/api/v1/auth/reset-password/confirm \
    -H "Content-Type: application/json" \
    -d '{
        "token":"invalid-token",
        "password":"newpass123",
        "confirm_password":"newpass123"
    }' \
    -w "\nHTTP Status: %{http_code}\n" \
    -s | jq .
echo "---"

# 4. Попытка с несовпадающими паролями
echo "4. Несовпадающие пароли:"
curl -X POST http://localhost:8000/api/v1/auth/reset-password/confirm \
    -H "Content-Type: application/json" \
    -d '{
        "token":"some-token",
        "password":"newpass123",
        "confirm_password":"different123"
    }' \
    -w "\nHTTP Status: %{http_code}\n" \
    -s | jq .
echo "---"

echo "Ожидается:"
echo "1. Письмо отправлено (mock)"
echo "2. Тот же ответ (безопасность)"
echo "3. Неверный токен - 400"
echo "4. Пароли не совпадают - 422"