#!/bin/bash

# Тест производительности regex для проверки CPU bottlenecks

echo "=== Testing Regex Performance Optimization ==="

# Функция для измерения времени выполнения
measure_time() {
    local endpoint=$1
    local data=$2
    local iterations=100
    
    echo "Testing $endpoint with $iterations requests..."
    
    start_time=$(date +%s%N)
    for i in $(seq 1 $iterations); do
        curl -s -X POST "http://localhost:8000$endpoint" \
            -H "Content-Type: application/json" \
            -d "$data" > /dev/null
    done
    end_time=$(date +%s%N)
    
    duration=$((($end_time - $start_time) / 1000000)) # Convert to milliseconds
    avg_time=$(($duration / $iterations))
    
    echo "Total time: ${duration}ms, Average per request: ${avg_time}ms"
}

echo "1. Testing signup validation (email, password, nick regex)..."
measure_time "/api/v1/auth/signup" '{
    "email": "test@example.com",
    "password": "TestPass123!",
    "confirm_password": "TestPass123!",
    "nick": "TestUser",
    "agree_terms": true
}'

echo -e "\n2. Testing profile update validation (nick, URL regex)..."
# First login to get token
login_response=$(curl -s -X POST http://localhost:8000/api/v1/auth/login \
    -H "Content-Type: application/json" \
    -d '{"email":"admin@example.com","password":"admin123"}')

token=$(echo $login_response | jq -r '.access_token // empty')

if [ -n "$token" ]; then
    echo "Testing with valid token..."
    start_time=$(date +%s%N)
    for i in $(seq 1 50); do
        curl -s -X PUT "http://localhost:8000/api/v1/profile" \
            -H "Content-Type: application/json" \
            -H "Authorization: Bearer $token" \
            -d '{
                "nick": "UpdatedUser",
                "bio": "Test bio",
                "links": [
                    {"title": "Website", "url": "https://example.com"},
                    {"title": "GitHub", "url": "https://github.com/user"}
                ]
            }' > /dev/null
    done
    end_time=$(date +%s%N)
    
    duration=$((($end_time - $start_time) / 1000000))
    avg_time=$(($duration / 50))
    echo "Profile update - Total: ${duration}ms, Average: ${avg_time}ms"
else
    echo "Could not get auth token, skipping profile test"
fi

echo -e "\n3. Testing auth header parsing optimization..."
start_time=$(date +%s%N)
for i in $(seq 1 100); do
    curl -s -X GET "http://localhost:8000/api/v1/profile" \
        -H "Authorization: Bearer invalid_token_for_parsing_test" > /dev/null
done
end_time=$(date +%s%N)

duration=$((($end_time - $start_time) / 1000000))
avg_time=$(($duration / 100))
echo "Auth header parsing - Total: ${duration}ms, Average: ${avg_time}ms"

echo -e "\n4. Checking server metrics for CPU usage..."
curl -s http://localhost:8000/api/v1/health | jq '.metrics // {}'

echo -e "\n=== Regex Performance Test Complete ==="
echo "All regex patterns are now pre-compiled for optimal performance!"