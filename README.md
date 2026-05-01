# Subscription Service

REST-сервис для агрегации данных об онлайн-подписках пользователей. Позволяет управлять записями о подписках и рассчитывать суммарные затраты за выбранные периоды.

## Технологический стек

*   **Language:** Go 1.23+
*   **Database:** PostgreSQL 16
*   **Query Builder:** [Squirrel](https://github.com/Masterminds/squirrel)
*   **Driver:** [pgx (v5)](https://github.com/jackc/pgx)
*   **Migrations:** [Goose](https://github.com/pressly/goose) (Embedded)
*   **Logging:** `slog` (Standard library)
*   **Architecture:** Clean Architecture (Entities, Repository, Service, Transport)
*   **Deployment:** Docker, Docker Compose

## Быстрый запуск

1. Клонируйте репозиторий:
```bash
git clone https://github.com/bigboynixxik/subs_service.git
cd subs_service
cp .env.example .env
```
> **Важно:** В `.env.example` уже преднастроены параметры для работы внутри Docker-сети. Если вы запускаете базу данных отдельно (не в Docker), измените `PG_DSN`.


2. Запустите проект через Docker Compose:
```Bash
docker-compose up --build
```

* Сервис автоматически применит миграции и поднимет HTTP-сервер на порту 8080.
## API Эндпоинты

### Подписки (Subscriptions)

| Метод | Эндпоинт | Описание |
| :--- | :--- | :--- |
| `POST` | `/subscriptions` | Создать новую запись о подписке |
| `GET` | `/subscriptions` | Получить список всех подписок |
| `GET` | `/subscriptions/{id}` | Получить детали конкретной подписки |
| `PUT` | `/subscriptions/{id}` | Обновить данные подписки |
| `DELETE` | `/subscriptions/{id}` | Удалить подписку |
| `GET` | `/subscriptions/cost` | Рассчитать стоимость за период |

## Архитектура проекта
Проект реализован согласно принципам чистой архитектуры:
- cmd/api: Точка входа в приложение.
- internal/app: Инициализация всех слоев и запуск сервера.
- internal/models: Бизнес-сущности.
- internal/service: Бизнес-логика.
- internal/repository: Работа с БД.
- internal/transport/rest: HTTP-хендлеры и маршрутизация.
- pkg: Общие библиотеки (logger, config, postgres pool, closer).

## Документация
Спецификация API доступна в формате Swagger в директории docs/.
```
docs/swagger.yaml
docs/swagger.json
```

### Примеры запросов

#### Создание подписки
```bash
curl -X POST http://localhost:8080/subscriptions \
-H "Content-Type: application/json" \
-d '{
"service_name": "Yandex Plus",
"price": 400,
"user_id": "60601fee-2bf1-4721-ae6f-7636e79a0cba",
"start_date": "07-2025"
}'
```

#### Расчет стоимости за период (Март - Май 2026)
```bash
curl "http://localhost:8080/subscriptions/cost?user_id=60601fee-2bf1-4721-ae6f-7636e79a0cba&from=03-2026&to=05-2026"
```
