.PHONY: test coverage coverage-html lint build run clean

# Установка зависимостей
deps:
	go mod download
	go mod tidy

# Сборка проекта
build:
	go build -o cmd/gophermart/gophermart cmd/gophermart/main.go

# Запуск приложения
run:
	./cmd/gophermart/gophermart

# Запуск тестов
test:
	go test -v -race ./...

# Покрытие тестами (с исключением моков)
coverage:
	@echo "Running tests with coverage (excluding mocks)..."
	@go test -coverprofile=coverage.out $$(go list ./... | grep -v '/mocks')
	@echo "\n=== Coverage (excluding mocks) ==="
	@go tool cover -func=coverage.out | tail -1

# Генерация HTML отчета покрытия
coverage-html: coverage
	@go tool cover -html=coverage.out -o coverage.html
	@echo "HTML coverage report: coverage.html"

# Запуск линтера
lint:
	golangci-lint run

# Генерация моков
mocks:
	mockery

# Очистка
clean:
	rm -f coverage.out coverage_filtered.out coverage.html
	rm -f cmd/gophermart/gophermart

# Запуск всех проверок
check: test lint
	@echo "All checks passed!"
