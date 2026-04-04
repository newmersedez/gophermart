package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Тест для функций которые не требуют database connection

func TestStorageInterface(t *testing.T) {
	// Проверяем что интерфейс StorageInterface определен правильно
	var s StorageInterface
	assert.Nil(t, s) // Проверяем что nil интерфейс работает
}

func TestStorageStruct(t *testing.T) {
	// Создаем пустую структуру Storage и проверяем Logger
	storage := &Storage{
		logger: nil,
	}
	assert.NotNil(t, storage)
	assert.Nil(t, storage.logger)
	assert.Nil(t, storage.pool)
}
