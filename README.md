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
```
2. Запустите проект через Docker Compose:
```Bash
docker-compose up --build
```

*Сервис автоматически применит миграции и поднимет HTTP-сервер на порту 8080.
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