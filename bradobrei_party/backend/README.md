# Bradobrei Party Backend

Бэкенд информационной системы сети барбершопов `Bradobrei Party` на `Go + Gin + GORM + PostgreSQL/PostGIS`.

## Что уже есть

- REST API для пользователей, сотрудников, салонов, услуг, материалов, бронирований, отзывов и отчётов.
- Автомиграция таблиц через GORM.
- Подключённый Swagger UI.
- JWT-аутентификация с `Bearer` токеном.
- Подготовка под PostGIS: `salons.location` хранится как `geometry(Point,4326)`.

## Структура

- `cmd/api/main.go` - точка входа, настройка БД, роутов, Swagger и middleware.
- `internal/models` - ORM-модели GORM.
- `internal/dto` - DTO запросов и ответов.
- `internal/handlers` - HTTP-эндпоинты.
- `internal/services` - прикладная логика.
- `internal/repository` - доступ к данным.
- `migrations` - SQL-миграции расширений и индексов.
- `docs` - сгенерированная Swagger-документация.

## Отчёты 2.2.x из ТЗ

- `2.2.1 Реестр персонала`
  - `GET /api/v1/reports/employees`
  - `internal/handlers/report_handler.go`
  - `internal/repository/report_repo.go`

- `2.2.2 Аналитический отчёт об операционной активности филиалов`
  - `GET /api/v1/reports/salon-activity`
  - основа данных также формируется через `POST /api/v1/bookings`

- `2.2.3 Статистика востребованности услуг`
  - `GET /api/v1/reports/service-popularity`

- `2.2.4 Отчёт о производительности и ресурсо-затратности мастеров`
  - `GET /api/v1/reports/master-activity`
  - рабочие данные для отчёта также участвуют в `GET /api/v1/bookings/master`

- `2.2.5 Журнал мониторинга качества обслуживания и обратной связи`
  - `GET /api/v1/reports/reviews`
  - `POST /api/v1/reviews`

- `2.2.6 Ведомость движения ТМЦ`
  - пока подготовлена модельная база: `Inventory`, `Material`, `ServiceMaterial`

- `2.2.7 Анализ клиентской лояльности и удержания`
  - пока подготовлена база на уровне модели `Booking`

- `2.2.8 Реестр отменённых и нереализованных бронирований`
  - пока подготовлена база на уровне `Booking.Status`

- `2.2.9 Сводный финансовый отчёт по статьям издержек`
  - пока не реализован

## Зависимости

Swagger-зависимости уже зафиксированы в `go.mod`, поэтому отдельно делать `go get github.com/swaggo/...` не нужно.

Чтобы скачать зависимости проекта:

```bash
go mod download
```

CLI `swag` нужен только если вы хотите заново сгенерировать папку `docs` после изменения аннотаций:

```bash
go install github.com/swaggo/swag/cmd/swag@latest
swag init -g ./cmd/api/main.go -o ./docs
```

## Настройка БД

Используется PostgreSQL с PostGIS.

Минимально нужно:

```sql
CREATE DATABASE bradobrei;
```

Приложение при старте само пытается выполнить:

```sql
CREATE SCHEMA IF NOT EXISTS public;
CREATE EXTENSION IF NOT EXISTS postgis;
```

Если база создавалась вручную и прав у пользователя недостаточно, выполните эти команды вручную под пользователем с нужными правами.

## Локальный запуск

Из папки `backend`:

```bash
go run ./cmd/api
```

По умолчанию сервер поднимается на:

```text
http://localhost:8080
```

## Swagger UI

Документация доступна по адресам:

```text
http://localhost:8080/swagger/index.html
http://localhost:8080/docs
```

Для защищённых эндпоинтов:

1. Выполните `POST /api/v1/auth/login`.
2. Скопируйте токен из ответа.
3. Нажмите кнопку `Authorize` в Swagger UI.
4. Вставьте значение в формате:

```text
Bearer <ваш_jwt_токен>
```

После этого Swagger будет автоматически подставлять заголовок `Authorization` в защищённые запросы.

## Примеры DTO в Swagger

Примеры для полей DTO задаются прямо в Go-тегах структуры:

```go
type LoginRequest struct {
    Username string `json:"username" binding:"required" example:"admin"`
    Password string `json:"password" binding:"required" example:"password"`
}
```

После изменения тегов нужно перегенерировать Swagger:

```bash
swag init -g ./cmd/api/main.go -o ./docs
```

Это удобно для локальной разработки:

- можно быстро логиниться тестовым пользователем через Swagger
- можно не собирать JSON вручную для `register`, `login`, `bookings`
- примеры сразу видны в schema и request body

Если у вас нет заранее созданного пользователя, сначала используйте `POST /api/v1/auth/register`, затем `POST /api/v1/auth/login`, затем `Authorize`.

## Сборка

Из папки `backend`:

```bash
go build -o bradobrei-api ./cmd/api
```

Запуск собранного бинарника:

```bash
./bradobrei-api
```

Для Git Bash в Windows:

```bash
./bradobrei-api.exe
```

## Полезные команды

Форматирование:

```bash
gofmt -w ./cmd ./internal
```

Проверка сборки:

```bash
go build ./...
```

Перегенерация Swagger:

```bash
swag init -g ./cmd/api/main.go -o ./docs
```

## Частые проблемы

### `schema for creating objects is not selected` / `SQLSTATE 3F000`

Обычно это значит одно из трёх:

- база пересоздана вручную и в ней нет схемы `public`
- у пользователя БД пустой `search_path`
- `.env` указывает не на тот экземпляр PostgreSQL

### `type "geometry" does not exist` / `SQLSTATE 42704`

Обычно это значит, что в текущей базе не включён `postgis`.

Решение:

```sql
CREATE EXTENSION IF NOT EXISTS postgis;
```

### Почему `go build` на Windows иногда падает, а `go run` идёт дальше

Иногда мешает доступ к кэшу Go в `%LOCALAPPDATA%\\go-build`.

Попробуйте:

```bash
go clean -cache
```
