package config

import (
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewConfig(t *testing.T) {
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	flag.CommandLine = flag.NewFlagSet("", flag.ContinueOnError)
	os.Args = []string{"test",
		"-a", "localhost:9999",
		"-d", "postgresql://localhost/test",
		"-r", "http://localhost:8081"}
	flag.CommandLine.Parse(os.Args[1:])

	cfg, err := NewConfig()

	require.NoError(t, err)
	require.NotNil(t, cfg)
	require.Equal(t, "localhost:9999", cfg.RunAddress)
	require.Equal(t, "postgresql://localhost/test", cfg.DatabaseURI)
	require.Equal(t, "http://localhost:8081", cfg.AccrualSystemAddress)
}

func TestRunAddressPriority(t *testing.T) {
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	tests := []struct {
		name           string
		envVars        map[string]string
		flagArgs       []string
		expectedResult string
	}{
		{
			name: "Flag has priority over environment",
			envVars: map[string]string{
				"RUN_ADDRESS": "localhost:9999",
			},
			flagArgs:       []string{"-a", "localhost:8888"},
			expectedResult: "localhost:8888",
		},
		{
			name: "Environment value if flag not set",
			envVars: map[string]string{
				"RUN_ADDRESS": "localhost:9999",
			},
			flagArgs:       []string{},
			expectedResult: "localhost:9999",
		},
		{
			name:           "Default value if both not set",
			envVars:        map[string]string{},
			flagArgs:       []string{},
			expectedResult: "localhost:8080",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for key := range tt.envVars {
				os.Unsetenv(key)
			}

			for key, value := range tt.envVars {
				t.Setenv(key, value)
			}

			if len(tt.flagArgs) > 0 {
				os.Args = append([]string{"cmd"}, tt.flagArgs...)
			} else {
				os.Args = []string{"cmd"}
			}

			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

			cfg, err := NewConfig()
			require.NoError(t, err)
			require.Equal(t, tt.expectedResult, cfg.RunAddress)
		})
	}
}

func TestDatabaseURIPriority(t *testing.T) {
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	tests := []struct {
		name           string
		envVars        map[string]string
		flagArgs       []string
		expectedResult string
	}{
		{
			name: "Flag has priority over environment",
			envVars: map[string]string{
				"DATABASE_URI": "postgresql://env/db",
			},
			flagArgs:       []string{"-d", "postgresql://flag/db"},
			expectedResult: "postgresql://flag/db",
		},
		{
			name: "Environment value if flag not set",
			envVars: map[string]string{
				"DATABASE_URI": "postgresql://env/db",
			},
			flagArgs:       []string{},
			expectedResult: "postgresql://env/db",
		},
		{
			name:           "Empty if both not set",
			envVars:        map[string]string{},
			flagArgs:       []string{},
			expectedResult: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for key := range tt.envVars {
				os.Unsetenv(key)
			}

			for key, value := range tt.envVars {
				t.Setenv(key, value)
			}

			if len(tt.flagArgs) > 0 {
				os.Args = append([]string{"cmd"}, tt.flagArgs...)
			} else {
				os.Args = []string{"cmd"}
			}

			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

			cfg, err := NewConfig()
			require.NoError(t, err)
			require.Equal(t, tt.expectedResult, cfg.DatabaseURI)
		})
	}
}

func TestAccrualSystemAddressPriority(t *testing.T) {
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	tests := []struct {
		name           string
		envVars        map[string]string
		flagArgs       []string
		expectedResult string
	}{
		{
			name: "Flag has priority over environment",
			envVars: map[string]string{
				"ACCRUAL_SYSTEM_ADDRESS": "http://localhost:8081",
			},
			flagArgs:       []string{"-r", "http://localhost:8081"},
			expectedResult: "http://localhost:8081",
		},
		{
			name: "Environment value if flag not set",
			envVars: map[string]string{
				"ACCRUAL_SYSTEM_ADDRESS": "http://localhost:8081",
			},
			flagArgs:       []string{},
			expectedResult: "http://localhost:8081",
		},
		{
			name:           "Empty if both not set",
			envVars:        map[string]string{},
			flagArgs:       []string{},
			expectedResult: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for key := range tt.envVars {
				os.Unsetenv(key)
			}

			for key, value := range tt.envVars {
				t.Setenv(key, value)
			}

			if len(tt.flagArgs) > 0 {
				os.Args = append([]string{"cmd"}, tt.flagArgs...)
			} else {
				os.Args = []string{"cmd"}
			}

			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

			cfg, err := NewConfig()
			require.NoError(t, err)
			require.Equal(t, tt.expectedResult, cfg.AccrualSystemAddress)
		})
	}
}

func TestNewConfig_EnvVarsAndFlags(t *testing.T) {
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	t.Setenv("RUN_ADDRESS", "env:9999")
	t.Setenv("DATABASE_URI", "postgresql://env/db")
	t.Setenv("ACCRUAL_SYSTEM_ADDRESS", "http://localhost:8081")

	os.Args = []string{"test", "-a", "flag:8888", "-d", "postgresql://flag/db"}
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	cfg, err := NewConfig()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	require.Equal(t, "flag:8888", cfg.RunAddress)
	require.Equal(t, "postgresql://flag/db", cfg.DatabaseURI)

	require.Equal(t, "http://localhost:8081", cfg.AccrualSystemAddress)
}

func TestNewConfig_DefaultValues(t *testing.T) {
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	os.Unsetenv("RUN_ADDRESS")
	os.Unsetenv("DATABASE_URI")
	os.Unsetenv("ACCRUAL_SYSTEM_ADDRESS")

	os.Args = []string{"test"}
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	cfg, err := NewConfig()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	require.Equal(t, "localhost:8080", cfg.RunAddress)

	require.Equal(t, "", cfg.DatabaseURI)
	require.Equal(t, "", cfg.AccrualSystemAddress)
}

func TestNewConfig_EnvironmentVariablesOnly(t *testing.T) {
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	t.Setenv("RUN_ADDRESS", "env:9999")
	t.Setenv("DATABASE_URI", "postgresql://env/db")
	t.Setenv("ACCRUAL_SYSTEM_ADDRESS", "http://localhost:8081")

	os.Args = []string{"test"}
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	cfg, err := NewConfig()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	require.Equal(t, "env:9999", cfg.RunAddress)
	require.Equal(t, "postgresql://env/db", cfg.DatabaseURI)
	require.Equal(t, "http://localhost:8081", cfg.AccrualSystemAddress)
}
