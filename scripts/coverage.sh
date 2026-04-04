#!/bin/bash

# Генерируем профиль покрытия, исключая mocks
go test -coverprofile=coverage.out $(go list ./... | grep -v '/mocks')

# Показываем статистику
echo "=== Coverage (excluding mocks) ==="
go tool cover -func=coverage.out | tail -1

# Опционально: генерируем HTML отчет
if [ "$1" == "html" ]; then
    go tool cover -html=coverage.out -o coverage.html
    echo "HTML report generated: coverage.html"
fi
