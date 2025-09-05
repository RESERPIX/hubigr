#!/bin/bash

# Тест connection pools для проверки memory leaks

echo "=== Testing Connection Pool Health ==="

# Проверяем health endpoint для статистики pools
echo "1. Checking pool statistics..."
curl -s http://localhost:8000/api/v1/health | jq '.database.pool_stats, .redis.pool_stats'

echo -e "\n2. Load testing to check for memory leaks..."

# Создаем нагрузку для проверки connection pools
for i in {1..50}; do
    # Параллельные запросы для создания нагрузки на connection pool
    curl -s -X POST http://localhost:8000/api/v1/auth/login \
        -H "Content-Type: application/json" \
        -d '{"email":"test@example.com","password":"wrongpass"}' > /dev/null &
    
    # Проверяем health каждые 10 запросов
    if [ $((i % 10)) -eq 0 ]; then
        echo "After $i requests:"
        curl -s http://localhost:8000/api/v1/health | jq -r '.database.pool_stats | "DB: acquired=\(.acquired_conns)/\(.max_conns), idle=\(.idle_conns)"'
        curl -s http://localhost:8000/api/v1/health | jq -r '.redis.pool_stats | "Redis: total=\(.total_conns), idle=\(.idle_conns), timeouts=\(.timeouts)"'
        echo "---"
    fi
done

# Ждем завершения всех запросов
wait

echo -e "\n3. Final pool statistics after load test:"
curl -s http://localhost:8000/api/v1/health | jq '.database.pool_stats, .redis.pool_stats'

echo -e "\n4. Checking for alerts..."
curl -s http://localhost:8000/api/v1/health | jq '.alerts'

echo -e "\n=== Connection Pool Test Complete ==="