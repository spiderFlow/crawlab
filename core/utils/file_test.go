package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExists(t *testing.T) {
	// Test cases
	pathString := "../config"
	wrongPathString := "test"

	// Test existing path
	res := Exists(pathString)
	assert.True(t, res, "Expected existing path to return true")

	// Test non-existing path
	wrongRes := Exists(wrongPathString)
	assert.False(t, wrongRes, "Expected non-existing path to return false")
}

func TestIsDir(t *testing.T) {
	// Test cases
	pathString := "../config"
	fileString := "../config/config.go"
	wrongString := "test"

	// Test directory path
	res := IsDir(pathString)
	assert.True(t, res, "Expected directory path to return true")

	// Test file path
	fileRes := IsDir(fileString)
	assert.False(t, fileRes, "Expected file path to return false")

	// Test non-existing path
	wrongRes := IsDir(wrongString)
	assert.False(t, wrongRes, "Expected non-existing path to return false")
}
