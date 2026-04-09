package main

import (
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRunWithoutDatabaseURI(t *testing.T) {
	originalURI := os.Getenv("DATABASE_URI")
	defer func() {
		if originalURI != "" {
			os.Setenv("DATABASE_URI", originalURI)
		} else {
			os.Unsetenv("DATABASE_URI")
		}
	}()

	flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)

	os.Unsetenv("DATABASE_URI")

	err := run()
	require.Error(t, err)
	require.EqualError(t, err, "database URI is required")
}

func TestRunWithInvalidDatabaseURI(t *testing.T) {
	originalURI := os.Getenv("DATABASE_URI")
	defer func() {
		if originalURI != "" {
			os.Setenv("DATABASE_URI", originalURI)
		} else {
			os.Unsetenv("DATABASE_URI")
		}
	}()

	flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)

	os.Setenv("DATABASE_URI", "invalid://uri")

	err := run()
	require.Error(t, err)
	require.ErrorContains(t, err, "failed to initialize storage")
}
