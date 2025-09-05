# Мониторинг и Алерты

## Запуск мониторинга

```bash
# Запуск основных сервисов
docker-compose up -d

# Запуск мониторинга
docker-compose -f docker-compose.monitoring.yml up -d
```

## Доступ к сервисам

- **Grafana**: http://localhost:3000 (admin/admin)
- **Prometheus**: http://localhost:9090
- **Alertmanager**: http://localhost:9093

## Метрики

### HTTP метрики
- `http_requests_total` - общее количество запросов
- `http_responses_total` - ответы по статус кодам
- `http_request_duration_ms` - средняя длительность

### Бизнес метрики
- `users_registered_total` - зарегистрированные пользователи
- `emails_sent_total` - отправленные email
- `login_attempts_total` - попытки входа
- `failed_logins_total` - неудачные входы

### Системные метрики
- `uptime_seconds` - время работы сервиса

## Алерты

### Автоматические алерты
- Высокое количество неудачных входов (>10)
- Недоступность базы данных
- Недоступность Redis
- Перезапуск сервиса

### Health Check
```bash
curl http://localhost:8000/api/v1/health
```

Возвращает статус всех компонентов и активные алерты.

## Настройка алертов

Алерты отправляются через webhook на `/api/v1/alerts/webhook` и логируются в structured формате для дальнейшей обработки внешними системами мониторинга.