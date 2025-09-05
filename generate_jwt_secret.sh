#!/bin/bash

# Генератор безопасного JWT секрета для Hubigr

echo "🔐 Генерация безопасного JWT секрета..."
echo ""

# Генерируем случайный ключ длиной 64 символа
JWT_SECRET=$(openssl rand -base64 48 | tr -d '\n')

echo "Ваш новый JWT секрет:"
echo "JWT_SECRET=$JWT_SECRET"
echo ""
echo "📋 Скопируйте эту строку в ваш .env файл"
echo "⚠️  НИКОГДА не делитесь этим секретом!"
echo ""
echo "Длина секрета: ${#JWT_SECRET} символов ✅"