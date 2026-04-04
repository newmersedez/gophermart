package main

import (
	"flag"
	"os"
	"testing"
)

func TestRunWithoutDatabaseURI(t *testing.T) {
	// Сохраним оригинальное значение
	originalURI := os.Getenv("DATABASE_URI")
	defer func() {
		if originalURI != "" {
			os.Setenv("DATABASE_URI", originalURI)
		} else {
			os.Unsetenv("DATABASE_URI")
		}
	}()

	// Сбросим флаги
	flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)

	// Убираем DATABASE_URI для теста
	os.Unsetenv("DATABASE_URI")

	err := run()
	if err == nil {
		t.Error("Expected error when DATABASE_URI is not set, got nil")
	}

	expectedMsg := "database URI is required"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message %q, got %q", expectedMsg, err.Error())
	}
}

func TestRunWithInvalidDatabaseURI(t *testing.T) {
	// Сохраним оригинальное значение
	originalURI := os.Getenv("DATABASE_URI")
	defer func() {
		if originalURI != "" {
			os.Setenv("DATABASE_URI", originalURI)
		} else {
			os.Unsetenv("DATABASE_URI")
		}
	}()

	// Сбросим флаги
	flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)

	// Устанавливаем невалидный DATABASE_URI
	os.Setenv("DATABASE_URI", "invalid://uri")

	err := run()
	if err == nil {
		t.Error("Expected error with invalid DATABASE_URI, got nil")
	}

	// Ожидаем ошибку инициализации хранилища
	expectedSubstring := "failed to initialize storage"
	if err.Error() == "" || len(err.Error()) < len(expectedSubstring) {
		t.Errorf("Expected error message to contain %q, got %q", expectedSubstring, err.Error())
	}
}
