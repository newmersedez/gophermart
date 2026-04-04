.PHONY: test coverage coverage-html lint build run clean

# Установка зависимостей
deps:
	go mod download
	go mod tidy

# Сборка проекта
build:
	go build -o cmd/gophermart/gophermart cmd/gophermart/main.go

# Запуск приложения
run: build
	./cmd/gophermart/gophermart

# Запуск тестов
test:
	go test -v -race ./...

# Покрытие тестами (с исключением моков)
coverage:
	@echo "Running tests with coverage..."
	@go test -coverprofile=coverage.out $$(go list ./... | grep -v '/mocks')
	@echo "\n=== Coverage ==="
	@go tool cover -func=coverage.out | tail -1

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
