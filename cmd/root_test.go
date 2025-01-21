package cmd

import (
	"bytes"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lugnut42/openipam/internal/config"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func captureOutput(f func()) string {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	f()
	log.SetOutput(os.Stderr) // restore default output
	return buf.String()
}

func setupTestConfig(t *testing.T) (string, func()) {
	tmpDir := t.TempDir()
	cfgFile := filepath.Join(tmpDir, "ipam-config.yaml")
	err := os.WriteFile(cfgFile, []byte("dataDir: "+tmpDir), 0644)
	require.NoError(t, err)

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return cfgFile, cleanup
}

func TestExecute(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	defer func() {
		os.Stdout = oldStdout
	}()

	cfgFile, cleanup := setupTestConfig(t)
	defer cleanup()

	tests := []struct {
		name        string
		args        []string
		envVars     map[string]string
		wantErr     bool
		errContains string
		setup       func()
		cleanup     func()
	}{
		{
			name:    "help command",
			args:    []string{"--help"},
			wantErr: false,
		},
		{
			name:    "valid config from flag",
			args:    []string{"--config", cfgFile},
			wantErr: false,
		},
		{
			name: "valid config from env var",
			envVars: map[string]string{
				"IPAM_CONFIG_PATH": cfgFile,
			},
			wantErr: false,
		},
		{
			name:        "invalid command",
			args:        []string{"invalidcmd"},
			wantErr:     true,
			errContains: "unknown command",
		},
		{
			name:        "no config specified",
			args:        []string{},
			wantErr:     true,
			errContains: "no configuration file specified",
		},
		{
			name:        "non-existent config file",
			args:        []string{"--config", "/nonexistent/config.yaml"},
			wantErr:     true,
			errContains: "configuration file not found",
		},
		{
			name: "invalid config file content",
			args: []string{"--config", cfgFile},
			setup: func() {
				os.WriteFile(cfgFile, []byte("invalid: yaml: content:"), 0644)
			},
			wantErr:     true,
			errContains: "error loading config file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset environment and config
			os.Clearenv()
			if tt.setup != nil {
				tt.setup()
			}

			// Set environment variables
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			// Create a new root command and set args
			oldArgs := os.Args
			os.Args = append([]string{"ipam"}, tt.args...)

			// Capture the error output
			var stderr bytes.Buffer
			log.SetOutput(&stderr)

			// Execute the command and capture result
			err := rootCmd.Execute()

			// Restore stderr
			log.SetOutput(os.Stderr)

			// Verify error conditions
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, stderr.String(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}

			// Cleanup
			os.Args = oldArgs
			for k := range tt.envVars {
				os.Unsetenv(k)
			}
			if tt.cleanup != nil {
				tt.cleanup()
			}
		})
	}

	w.Close()
	output, _ := io.ReadAll(r)
	assert.NotEmpty(t, string(output))
}

func TestPersistentPreRunE(t *testing.T) {
	cfgFile, cleanup := setupTestConfig(t)
	defer cleanup()

	tests := []struct {
		name        string
		cmd         *cobra.Command
		args        []string
		envVars     map[string]string
		setupConfig string
		wantErr     bool
		errContains string
	}{
		{
			name: "config init command skips check",
			cmd: func() *cobra.Command {
				parent := &cobra.Command{Use: "config"}
				cmd := &cobra.Command{Use: "init"}
				parent.AddCommand(cmd)
				return cmd
			}(),
			wantErr: false,
		},
		{
			name:    "valid config with debug logging",
			cmd:     &cobra.Command{Use: "test"},
			args:    []string{"--config", cfgFile},
			wantErr: false,
		},
		{
			name: "env var config with logging",
			cmd:  &cobra.Command{Use: "test"},
			envVars: map[string]string{
				"IPAM_CONFIG_PATH": cfgFile,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset environment
			os.Clearenv()

			// Set environment variables
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			// Capture logs
			output := captureOutput(func() {
				err := rootCmd.PersistentPreRunE(tt.cmd, tt.args)
				if tt.wantErr {
					assert.Error(t, err)
					if tt.errContains != "" {
						assert.Contains(t, err.Error(), tt.errContains)
					}
				} else {
					assert.NoError(t, err)
				}
			})

			// Verify debug logging
			assert.True(t, strings.Contains(output, "DEBUG:"))

			// Cleanup
			for k := range tt.envVars {
				os.Unsetenv(k)
			}
		})
	}
}

func TestRootCmdInitialization(t *testing.T) {
	tests := []struct {
		name string
		want struct {
			configFlag bool
			configType string
		}
	}{
		{
			name: "default initialization",
			want: struct {
				configFlag bool
				configType string
			}{
				configFlag: true,
				configType: "string",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset the root command for testing
			cfg = nil
			rootCmd.ResetFlags()

			output := captureOutput(func() {
				// Call the package init function indirectly by reinitializing flags
				cfg = &config.Config{}
				rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "Path to configuration file")
			})

			// Verify config initialization
			assert.NotNil(t, cfg)

			// Verify flag registration
			flag := rootCmd.PersistentFlags().Lookup("config")
			assert.NotNil(t, flag)
			assert.Equal(t, tt.want.configType, flag.Value.Type())
			assert.Equal(t, "Path to configuration file", flag.Usage)

			// Verify debug logging output contains initialization messages
			assert.Contains(t, output, "DEBUG:")
		})
	}
}
