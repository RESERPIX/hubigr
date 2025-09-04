#!/bin/bash

# Тест списка сабмитов пользователя согласно ТЗ UC-1.2.2

echo "Тестирование списка сабмитов пользователя..."

# 1. Попытка получить сабмиты без авторизации
echo "1. Запрос без авторизации:"
curl -X GET http://localhost:8000/api/v1/profile/submissions \
    -w "\nHTTP Status: %{http_code}\n" \
    -s | jq .
echo "---"

# 2. Регистрация и получение токена
echo "2. Регистрация пользователя:"
SIGNUP_RESPONSE=$(curl -X POST http://localhost:8000/api/v1/auth/signup \
    -H "Content-Type: application/json" \
    -d '{
        "email": "testuser@example.com",
        "password": "testpass123",
        "confirm_password": "testpass123",
        "nick": "testuser",
        "agree_terms": true
    }' -s)

echo $SIGNUP_RESPONSE | jq .

# Извлекаем user_id для тестовых данных
USER_ID=$(echo $SIGNUP_RESPONSE | jq -r '.user_id')
echo "User ID: $USER_ID"
echo "---"

# 3. Получение сабмитов с пагинацией (пока будет пустой список)
echo "3. Получение сабмитов (требует JWT токен):"
echo "Примечание: Для полного теста нужен JWT токен после подтверждения email"
echo "---"

echo "Ожидается:"
echo "1. 401 - Unauthorized (нет токена)"
echo "2. Успешная регистрация"
echo "3. Для получения сабмитов нужен JWT токен"
echo ""
echo "После подтверждения email и получения токена:"
echo "curl -H 'Authorization: Bearer <token>' http://localhost:8000/api/v1/profile/submissions"