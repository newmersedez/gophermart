# Gophermart - Накопительная система лояльности

[![Tests & Coverage](https://github.com/newmersedez/gophermart/actions/workflows/coverage.yml/badge.svg)](https://github.com/newmersedez/gophermart/actions/workflows/coverage.yml)
[![golangci-lint](https://github.com/newmersedez/gophermart/actions/workflows/golangci-lint.yml/badge.svg)](https://github.com/newmersedez/gophermart/actions/workflows/golangci-lint.yml)

## 📊 Coverage: AUTO_COVERAGE% AUTO_STATUS

Дипломный проект курса «Go-разработчик» от Яндекс Практикум.

## Описание

Gophermart — это система лояльности для интернет-магазина «Гофермарт». Система позволяет пользователям:
- Регистрироваться и аутентифицироваться
- Загружать номера заказов для начисления баллов
- Просматривать баланс и историю начислений
- Списывать баллы для оплаты заказов

## Требования

- Go 1.23 или выше
- PostgreSQL 12 или выше
- Docker (опционально)

## Локальная установка и запуск

### 1. Клонирование репозитория

```bash
git clone <repository_url>
cd gophermart
```

### 2. Установка зависимостей

```bash
go mod download
```

### 3. Настройка PostgreSQL

Создайте две базы данных - для gophermart и accrual:

```bash
createdb gophermart
createdb accrual
```

### 4. Сборка проекта

```bash
go build -o cmd/gophermart/gophermart cmd/gophermart/main.go
```

### 5. Запуск сервиса

#### С помощью флагов:

```bash
./cmd/gophermart/gophermart \
    -a=localhost:8080 \
    -d="postgresql://postgres:postgres@localhost:5432/gophermart?sslmode=disable" \
    -r=http://localhost:8081
```

#### С помощью переменных окружения:

```bash
export RUN_ADDRESS="localhost:8080"
export DATABASE_URI="postgresql://postgres:postgres@localhost:5432/gophermart?sslmode=disable"
export ACCRUAL_SYSTEM_ADDRESS="http://localhost:8081"
./cmd/gophermart/gophermart
```

**Примечание:** Флаги имеют приоритет над переменными окружения.

### Параметры запуска

- `-a` / `RUN_ADDRESS` - адрес и порт для запуска сервиса (по умолчанию: `localhost:8080`)
- `-d` / `DATABASE_URI` - URI подключения к PostgreSQL (обязательный параметр)
- `-r` / `ACCRUAL_SYSTEM_ADDRESS` - адрес системы расчёта баллов лояльности (опционально)

## API Endpoints

### Регистрация и аутентификация

- `POST /api/user/register` - регистрация нового пользователя
- `POST /api/user/login` - аутентификация пользователя

### Работа с заказами

- `POST /api/user/orders` - загрузка номера заказа для расчёта
- `GET /api/user/orders` - получение списка загруженных заказов

### Баланс и списания

- `GET /api/user/balance` - получение баланса счёта
- `POST /api/user/balance/withdraw` - списание баллов
- `GET /api/user/withdrawals` - получение истории списаний

## Запуск автотестов

```bash
gophermarttest \
    -test.v \
    -test.run=^TestGophermart$ \
    -gophermart-binary-path=./cmd/gophermart/gophermart \
    -gophermart-host=localhost \
    -gophermart-port=8080 \
    -gophermart-database-uri="postgresql://postgres:postgres@localhost:5432/gophermart?sslmode=disable" \
    -accrual-binary-path=./cmd/accrual/accrual_darwin_arm64 \
    -accrual-host=localhost \
    -accrual-port=8081 \
    -accrual-database-uri="postgresql://postgres:postgres@localhost:5432/accrual?sslmode=disable"
```

**Примечание:** Выберите нужный бинарник accrual для вашей платформы:
- `accrual_darwin_amd64` - macOS Intel
- `accrual_darwin_arm64` - macOS Apple Silicon
- `accrual_linux_amd64` - Linux
- `accrual_windows_amd64.exe` - Windows

## Тестирование и покрытие кода

### Запуск unit-тестов

```bash
# Запустить все тесты
make test

# Или напрямую через go
go test -v -race ./...
```

### Проверка покрытия тестами

```bash
# Покрытие в консоли (исключая моки)
make coverage

**Примечание:** Все команды автоматически исключают папки `mocks` из расчета покрытия.

### Интеграционные тесты storage

Тесты для пакета `internal/storage` требуют запущенную базу данных PostgreSQL. Локально они будут пропущены, если база данных недоступна.

Для запуска интеграционных тестов:

```bash
# Запустите PostgreSQL через Docker
docker run -d \
  --name gophermart-test-db \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=test \
  -e POSTGRES_DB=gophermart_test \
  -p 5432:5432 \
  postgres:15

# Установите переменную окружения
export TEST_DATABASE_URI="postgres://postgres:test@localhost:5432/gophermart_test?sslmode=disable"

# Запустите тесты
go test -v ./internal/storage
```

В GitHub Actions интеграционные тесты запускаются автоматически с PostgreSQL service container.

### Доступные Make команды

- `make deps` - установка зависимостей
- `make build` - сборка проекта
- `make test` - запуск тестов
- `make coverage` - проверка покрытия тестами
- `make coverage-html` - генерация HTML отчета покрытия
- `make lint` - проверка кода линтером
- `make mocks` - генерация моков через mockery
- `make check` - запуск всех проверок (тесты + линтер)
- `make clean` - очистка артефактов

## Архитектура проекта

```
gophermart/
├── cmd/
│   └── gophermart/          # Точка входа приложения
├── internal/
│   ├── accrual/             # Клиент для системы начисления баллов
│   ├── auth/                # Аутентификация и авторизация
│   ├── config/              # Конфигурация приложения
│   ├── handlers/            # HTTP обработчики
│   ├── middleware/          # Middleware (логирование, сжатие, авторизация)
│   ├── models/              # Модели данных
│   ├── storage/             # Работа с БД и миграции
│   ├── utils/               # Утилиты (алгоритм Луна)
│   └── worker/              # Фоновые процессы для обработки заказов
└── README.md
```

## Особенности реализации

- Слоистая архитектура с разделением ответственности
- JWT токены для аутентификации (HttpOnly cookies)
- Хеширование паролей с использованием bcrypt
- Миграции базы данных через golang-migrate
- Проверка номеров заказов по алгоритму Луна
- Graceful shutdown для корректного завершения работы
- Middleware для логирования запросов и gzip сжатия
- Фоновый воркер для обработки заказов через систему accrual

## CI/CD

Проект использует GitHub Actions:
- `golangci-lint.yml` - линтер кода
- `statictest.yml` - статический анализ от Яндекс Практикум
- `gophermart.yml` - автотесты проекта

## Лицензия

Образовательный проект Яндекс Практикум