#!/bin/bash

# Тест защиты от Path Traversal атак

BASE_URL="http://localhost:8000"

echo "🔒 Тестирование защиты от Path Traversal..."
echo ""

# Тест 1: Попытка доступа к /etc/passwd
echo "1. Попытка доступа к /etc/passwd:"
curl -s -w "Status: %{http_code}\n" "$BASE_URL/uploads/../../../etc/passwd"
echo ""

# Тест 2: Попытка доступа к системным файлам
echo "2. Попытка доступа к системным файлам:"
curl -s -w "Status: %{http_code}\n" "$BASE_URL/uploads/../../../../../../etc/hosts"
echo ""

# Тест 3: Попытка доступа через URL encoding
echo "3. Попытка через URL encoding:"
curl -s -w "Status: %{http_code}\n" "$BASE_URL/uploads/%2e%2e%2f%2e%2e%2fetc%2fpasswd"
echo ""

# Тест 4: Попытка доступа к исходному коду
echo "4. Попытка доступа к исходному коду:"
curl -s -w "Status: %{http_code}\n" "$BASE_URL/uploads/../main.go"
echo ""

# Тест 5: Попытка доступа к .env файлу
echo "5. Попытка доступа к .env файлу:"
curl -s -w "Status: %{http_code}\n" "$BASE_URL/uploads/../.env"
echo ""

# Тест 6: Нормальный доступ к файлу (должен работать)
echo "6. Нормальный доступ к файлу:"
curl -s -w "Status: %{http_code}\n" "$BASE_URL/uploads/avatars/test.jpg"
echo ""

echo "✅ Все запросы должны возвращать 403/404, кроме последнего"
echo "❌ Если какой-то запрос вернул 200 - есть уязвимость!"