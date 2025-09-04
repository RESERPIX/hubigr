#!/bin/bash

# Тест rate limiting согласно ТЗ: 5 попыток входа/мин

echo "Тестирование rate limiting для входа..."

# Запуск 6 попыток входа подряд
for i in {1..6}; do
    echo "Попытка $i:"
    curl -X POST http://localhost:8000/api/v1/auth/login \
        -H "Content-Type: application/json" \
        -d '{"email":"test@example.com","password":"wrongpass"}' \
        -w "\nHTTP Status: %{http_code}\n" \
        -s | jq .
    echo "---"
done

echo "Ожидается: первые 5 попыток - 401, 6-я попытка - 429 (rate limit exceeded)"