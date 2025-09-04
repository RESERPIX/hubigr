#!/bin/bash

# Тест email verification согласно ТЗ UC-1.1.1

echo "Тестирование email verification..."

# 1. Регистрация пользователя
echo "1. Регистрация пользователя:"
SIGNUP_RESPONSE=$(curl -X POST http://localhost:8000/api/v1/auth/signup \
    -H "Content-Type: application/json" \
    -d '{
        "email": "test@example.com",
        "password": "testpass123",
        "confirm_password": "testpass123",
        "nick": "testuser",
        "agree_terms": true
    }' -s)

echo $SIGNUP_RESPONSE | jq .
echo "---"

# 2. Попытка входа без подтверждения email
echo "2. Попытка входа без подтверждения email:"
curl -X POST http://localhost:8000/api/v1/auth/login \
    -H "Content-Type: application/json" \
    -d '{"email":"test@example.com","password":"testpass123"}' \
    -w "\nHTTP Status: %{http_code}\n" \
    -s | jq .
echo "---"

# 3. Повторная отправка письма подтверждения
echo "3. Повторная отправка письма подтверждения:"
curl -X POST http://localhost:8000/api/v1/auth/resend-verification \
    -H "Content-Type: application/json" \
    -d '{"email":"test@example.com"}' \
    -w "\nHTTP Status: %{http_code}\n" \
    -s | jq .
echo "---"

echo "Ожидается:"
echo "1. Регистрация успешна, письмо отправлено (mock)"
echo "2. Вход заблокирован - email_not_verified"
echo "3. Новое письмо отправлено"