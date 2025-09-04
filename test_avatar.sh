#!/bin/bash

# Тест загрузки аватара согласно ТЗ UC-1.2.1

echo "Тестирование загрузки аватара..."

# Создаем тестовое изображение (1x1 PNG)
echo "Создание тестового изображения..."
echo -n -e '\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR\x00\x00\x00\x01\x00\x00\x00\x01\x08\x02\x00\x00\x00\x90wS\xde\x00\x00\x00\tpHYs\x00\x00\x0b\x13\x00\x00\x0b\x13\x01\x00\x9a\x9c\x18\x00\x00\x00\nIDATx\x9cc\xf8\x00\x00\x00\x01\x00\x01\x00\x00\x00\x00IEND\xaeB`\x82' > test_avatar.png

# 1. Попытка загрузки без авторизации
echo "1. Загрузка без авторизации:"
curl -X POST http://localhost:8000/api/v1/profile/avatar \
    -F "avatar=@test_avatar.png" \
    -w "\nHTTP Status: %{http_code}\n" \
    -s | jq .
echo "---"

# 2. Попытка загрузки без файла
echo "2. Запрос без файла:"
curl -X POST http://localhost:8000/api/v1/profile/avatar \
    -H "Authorization: Bearer fake-token" \
    -w "\nHTTP Status: %{http_code}\n" \
    -s | jq .
echo "---"

# 3. Создание большого файла (>2MB)
echo "3. Создание файла >2MB:"
dd if=/dev/zero of=large_avatar.png bs=1M count=3 2>/dev/null
curl -X POST http://localhost:8000/api/v1/profile/avatar \
    -H "Authorization: Bearer fake-token" \
    -F "avatar=@large_avatar.png" \
    -w "\nHTTP Status: %{http_code}\n" \
    -s | jq .
echo "---"

# Очистка
rm -f test_avatar.png large_avatar.png

echo "Ожидается:"
echo "1. 401 - Unauthorized (нет токена)"
echo "2. 400 - Bad Request (нет файла)"
echo "3. 422 - Validation Error (файл слишком большой)"
echo ""
echo "Для полного теста нужен JWT токен:"
echo "curl -H 'Authorization: Bearer <token>' -F 'avatar=@image.png' http://localhost:8000/api/v1/profile/avatar"