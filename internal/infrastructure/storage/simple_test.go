package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)



func TestStorageInterface(t *testing.T) {
	
	var s StorageInterface
	assert.Nil(t, s) 
}

func TestStorageStruct(t *testing.T) {
	
	storage := &Storage{
		logger: nil,
	}
	assert.NotNil(t, storage)
	assert.Nil(t, storage.logger)
	assert.Nil(t, storage.pool)
}
