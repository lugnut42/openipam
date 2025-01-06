package cmd

import (
	"log"
	"os"
	"path/filepath"
	"testing"

	"openipam/internal/ipam/config"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestBlockCommands(t *testing.T) {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Llongfile)

	tempDir := t.TempDir()
	log.Printf("DEBUG: Created temp directory at: %s", tempDir)

	configFilePath := filepath.Join(tempDir, "ipam-config.yaml")
	blockFilePath := filepath.Join(tempDir, "ip-blocks.yaml")
	log.Printf("DEBUG: Config file path: %s", configFilePath)
	log.Printf("DEBUG: Block file path: %s", blockFilePath)

	initConfig := func() {
		log.Printf("DEBUG: Starting initConfig()")

		// Create a new config instance
		cfg = &config.Config{
			BlockFiles: map[string]string{"default": blockFilePath},
			ConfigFile: configFilePath,
		}

		rootCmd := &cobra.Command{Use: "ipam"}
		rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "Path to configuration file")
		rootCmd.AddCommand(configCmd)

		log.Printf("DEBUG: About to execute config init command")
		initArgs := []string{"config", "init", "--config", configFilePath, "--block-yaml-file", blockFilePath}
		log.Printf("DEBUG: Config init args: %v", initArgs)
		rootCmd.SetArgs(initArgs)

		err := rootCmd.Execute()
		if err != nil {
			log.Printf("ERROR: Failed to execute config init command: %v", err)
		}
		assert.NoError(t, err)

		// Create a block YAML file with valid initial structure
		log.Printf("DEBUG: Writing initial block file content")
		initialBlockContent := `[]`
		err = os.WriteFile(blockFilePath, []byte(initialBlockContent), 0644)
		if err != nil {
			log.Printf("ERROR: Failed to write initial block file: %v", err)
		}
		assert.NoError(t, err)

		// Debug prints
		configContent, err := os.ReadFile(configFilePath)
		if err != nil {
			log.Printf("ERROR: Failed to read config file: %v", err)
		} else {
			log.Printf("DEBUG: Config file contents:\n%s", string(configContent))
		}
		assert.NoError(t, err)

		blockContent, err := os.ReadFile(blockFilePath)
		if err != nil {
			log.Printf("ERROR: Failed to read block file: %v", err)
		} else {
			log.Printf("DEBUG: Block file contents:\n%s", string(blockContent))
		}
		assert.NoError(t, err)

		log.Printf("DEBUG: Completed initConfig()")
	}

	initConfig()

	executeCommand := func(args ...string) error {
		log.Printf("DEBUG: Executing command with args: %v", args)
		rootCmd := &cobra.Command{Use: "ipam"}

		// Important: Set up the persistent pre-run hook
		rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
			log.Printf("DEBUG: PreRun hook - Loading config from: %s", configFilePath)
			var err error
			cfg, err = config.LoadConfig(configFilePath)
			if err != nil {
				log.Printf("ERROR: Failed to load config in PreRun: %v", err)
				return err
			}
			log.Printf("DEBUG: PreRun hook - Loaded config: %+v", cfg)
			return nil
		}

		rootCmd.PersistentFlags().StringVar(&cfgFile, "config", configFilePath, "Path to configuration file")
		rootCmd.AddCommand(blockCmd)
		rootCmd.SetArgs(args)

		err := rootCmd.Execute()
		if err != nil {
			log.Printf("ERROR: Command execution failed: %v", err)
		} else {
			log.Printf("DEBUG: Command executed successfully")
		}
		return err
	}

	t.Run("block add", func(t *testing.T) {
		log.Printf("DEBUG: Starting 'block add' test")

		// Add additional validation
		_, err := os.Stat(configFilePath)
		assert.NoError(t, err, "Config file should exist")

		_, err = os.Stat(blockFilePath)
		assert.NoError(t, err, "Block file should exist")

		configContent, err := os.ReadFile(configFilePath)
		assert.NoError(t, err)
		log.Printf("DEBUG: Config file contents before block add:\n%s", string(configContent))

		err = executeCommand("block", "add", "--cidr", "10.0.0.0/16", "--description", "Test Block", "--file", "default")
		assert.NoError(t, err)

		// Verify block was added
		blockContent, err := os.ReadFile(blockFilePath)
		assert.NoError(t, err)
		log.Printf("DEBUG: Block file contents after add:\n%s", string(blockContent))
	})

	// Rest of the tests...
}
