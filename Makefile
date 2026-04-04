.PHONY: test coverage coverage-html coverage-check coverage-badge lint build run clean deps mocks

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

# HTML отчет покрытия
coverage-html: coverage
	@echo "Generating HTML coverage report..."
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"
	@echo "Open in browser: open coverage.html"

# Быстрая проверка покрытия
coverage-check: coverage
	@COVERAGE=$$(go tool cover -func=coverage.out | grep total | awk '{print $$3}' | sed 's/%//'); \
	echo "Current coverage: $$COVERAGE%"; \
	if [ $$(echo "$$COVERAGE >= 60" | bc -l) -eq 1 ]; then \
		echo "✅ Coverage meets target (≥60%)"; \
	else \
		echo "❌ Coverage below target (60%)"; \
		exit 1; \
	fi

# Обновить badge покрытия локально
coverage-badge: coverage
	@echo "Updating coverage badge..."
	@COVERAGE=$$(go tool cover -func=coverage.out | grep total | awk '{print $$3}' | sed 's/%//'); \
	COLOR="red"; \
	if [ $$(echo "$$COVERAGE > 80" | bc -l) -eq 1 ]; then \
		COLOR="brightgreen"; \
	elif [ $$(echo "$$COVERAGE > 60" | bc -l) -eq 1 ]; then \
		COLOR="yellow"; \
	fi; \
	mkdir -p .github/badges; \
	echo "{\"schemaVersion\": 1, \"label\": \"coverage\", \"message\": \"$${COVERAGE}%\", \"color\": \"$$COLOR\"}" > .github/badges/coverage.json; \
	echo "✅ Coverage badge updated: $$COVERAGE% ($$COLOR)"; \
	echo "Now run: git add .github/badges/coverage.json && git commit -m \"Update coverage: $$COVERAGE%\" && git push"

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
