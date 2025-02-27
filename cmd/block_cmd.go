package cmd

import (
	"fmt"
	"os"

	"github.com/lugnut42/openipam/internal/ipam"

	"github.com/spf13/cobra"
)

// blockCmd represents the block command
var blockCmd = &cobra.Command{
	Use:   "block",
	Short: "Manage IP blocks",
	Long:  `Create, list, show, and delete IP blocks.`,
}

var blockListCmd = &cobra.Command{
	Use:   "list",
	Short: "List IP blocks",
	Long:  `List all available IP blocks.`,
	Run: func(cmd *cobra.Command, args []string) {
		fileKey, _ := cmd.Flags().GetString("file")
		err := ipam.ListBlocks(cfg, fileKey)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
	},
}

var blockCreateCmd = &cobra.Command{
	//var blockCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new IP block",
	Long:  `Create a new IP block to the YAML file.`,
	Run: func(cmd *cobra.Command, args []string) {
		cidr, _ := cmd.Flags().GetString("cidr")
		description, _ := cmd.Flags().GetString("description")
		fileKey, _ := cmd.Flags().GetString("file")

		err := ipam.AddBlock(cfg, cidr, description, fileKey)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}

		fmt.Println("Block added successfully!")
	},
}

var blockShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show details of an IP block",
	Long:  `Display details of a specific IP block.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cidr := args[0]
		fileKey, _ := cmd.Flags().GetString("file")

		err := ipam.ShowBlock(cfg, cidr, fileKey)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
	},
}

var blockDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete an IP block",
	Long:  `Delete a specific IP block from the YAML file.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cidr := args[0]
		force, _ := cmd.Flags().GetBool("force")
		fileKey, _ := cmd.Flags().GetString("file")

		// Pass fileKey to DeleteBlock 
		err := ipam.DeleteBlock(cfg, cidr, force, fileKey)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}

		fmt.Println("Block deleted successfully!")
	},
}

var blockAvailableCmd = &cobra.Command{
	Use:   "available",
	Short: "List available CIDR ranges within a block",
	Long:  `List all available CIDR ranges within a specified block.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cidr := args[0]
		fileKey, _ := cmd.Flags().GetString("file")

		err := ipam.ListAvailableCIDRs(cfg, cidr, fileKey)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
	},
}

// blockValidateCmd represents the validate command
var blockValidateCmd = &cobra.Command{
	Use:   "validate [file-key]",
	Short: "Validate block file integrity",
	Long: `Validate the integrity and consistency of block configuration files.

This command performs comprehensive validation on your block configuration files, checking:
- File structure and format correctness
- CIDR syntax validation for blocks and subnets
- Detection of duplicate names, CIDRs, and other resources
- Subnet containment within parent blocks
- Network overlap detection
- Cross-reference integrity with patterns and other components
- Required field presence

If a file-key is provided, it validates only that specific block file.
Without a file-key, it validates the default block file.`,
	Run: func(cmd *cobra.Command, args []string) {
		fileKey := "default"
		if len(args) > 0 {
			fileKey = args[0]
		}
		
		results, err := ipam.ValidateBlockFile(cfg, fileKey)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
		
		fmt.Printf("=== Validating Block File: %s ===\n", fileKey)
		ipam.PrintValidationResults(results)
		
		if results.ErrorCount > 0 {
			os.Exit(1)
		}
	},
}

// blockUtilCommand represents the utilization command
var blockUtilCommand = &cobra.Command{
	Use:   "utilization [CIDR]",
	Short: "Show IP utilization statistics for a block",
	Long:  `Display IP address utilization statistics for a specific block or all blocks.
	
When used with a CIDR argument, shows detailed utilization for that specific block.
When used with the --all flag, shows utilization summary for all blocks in the file.`,
	Run: func(cmd *cobra.Command, args []string) {
		fileKey, _ := cmd.Flags().GetString("file")
		all, _ := cmd.Flags().GetBool("all")

		var err error
		if all {
			err = ipam.PrintAllBlocksUtilization(cfg, fileKey)
		} else if len(args) == 1 {
			cidr := args[0]
			err = ipam.PrintBlockUtilization(cfg, cidr, fileKey)
		} else {
			fmt.Fprintln(os.Stderr, "Error: Either specify a CIDR or use the --all flag")
			os.Exit(1)
		}

		if err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
	},
}


func init() {
	rootCmd.AddCommand(blockCmd)
	
	// List command
	blockCmd.AddCommand(blockListCmd)
	blockListCmd.Flags().StringP("file", "f", "default", "Key for the block file in the configuration (default is 'default')")
	
	// Create command
	blockCreateCmd.Flags().StringP("cidr", "c", "", "CIDR block (required)")
	if err := blockCreateCmd.MarkFlagRequired("cidr"); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
	blockCreateCmd.Flags().StringP("description", "d", "", "Description of the block")
	blockCreateCmd.Flags().StringP("file", "f", "default", "Key for the block file in the configuration (default is 'default')")
	blockCmd.AddCommand(blockCreateCmd)
	
	// Show command
	blockShowCmd.Flags().StringP("file", "f", "default", "Key for the block file in the configuration (default is 'default')")
	blockCmd.AddCommand(blockShowCmd)
	
	// Delete command
	blockDeleteCmd.Flags().BoolP("force", "", false, "Force deletion without confirmation")
	blockDeleteCmd.Flags().StringP("file", "f", "default", "Key for the block file in the configuration (default is 'default')")
	blockCmd.AddCommand(blockDeleteCmd)
	
	// Available command
	blockAvailableCmd.Flags().StringP("file", "f", "default", "Key for the block file in the configuration (default is 'default')")
	blockCmd.AddCommand(blockAvailableCmd)
	
	// Validate command
	blockValidateCmd.Flags().StringP("file", "f", "default", "Key for the block file in the configuration (default is 'default')")
	blockCmd.AddCommand(blockValidateCmd)
	
	// Utilization command
	blockUtilCommand.Flags().StringP("file", "f", "default", "Key for the block file in the configuration (default is 'default')")
	blockUtilCommand.Flags().BoolP("all", "a", false, "Show utilization for all blocks")
	blockCmd.AddCommand(blockUtilCommand)
}
