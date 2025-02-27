package cmd

import (
	"fmt"
	"os"

	"github.com/lugnut42/openipam/internal/ipam"
	"github.com/spf13/cobra"
)

var blockCmd = &cobra.Command{
	Use:   "block",
	Short: "Manage IP address blocks",
	Long:  `Add, list, show, and delete IP address blocks.`,
}

// blockCreateCmd represents the create command
var blockCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new IP address block",
	Long: `Create a new IP address block with a specified CIDR range.
	
Example:
  ipam block create --cidr 10.0.0.0/16 --description "Production Network" --file prod`,
	Run: func(cmd *cobra.Command, args []string) {
		cidr, _ := cmd.Flags().GetString("cidr")
		description, _ := cmd.Flags().GetString("description")
		fileKey, _ := cmd.Flags().GetString("file")

		err := ipam.AddBlock(cfg, cidr, description, fileKey)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}

		fmt.Printf("Created block %s in %s file\n", cidr, fileKey)
	},
}

// blockListCmd represents the list command
var blockListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all IP address blocks",
	Long: `List all IP address blocks in the block file.
	
Example:
  ipam block list
  ipam block list --file prod`,
	Run: func(cmd *cobra.Command, args []string) {
		fileKey, _ := cmd.Flags().GetString("file")

		err := ipam.ListBlocks(cfg, fileKey)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
	},
}

// blockShowCmd represents the show command
var blockShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show details of an IP address block",
	Long: `Show details of an IP address block, including its CIDR, description, and subnets.
	
Example:
  ipam block show 10.0.0.0/16
  ipam block show 10.0.0.0/16 --file prod`,
	Args: cobra.ExactArgs(1),
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

// blockDeleteCmd represents the delete command
var blockDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete an IP address block",
	Long: `Delete an IP address block.
	
Example:
  ipam block delete 10.0.0.0/16
  ipam block delete 10.0.0.0/16 --force
  ipam block delete 10.0.0.0/16 --file prod`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cidr := args[0]
		force, _ := cmd.Flags().GetBool("force")
		fileKey, _ := cmd.Flags().GetString("file")

		err := ipam.DeleteBlock(cfg, cidr, force, fileKey)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}

		fmt.Printf("Deleted block %s from %s file\n", cidr, fileKey)
	},
}

// blockAvailableCmd represents the available command
var blockAvailableCmd = &cobra.Command{
	Use:   "available",
	Short: "Show available subnets in a block",
	Long: `Show the available subnets in a block that can be allocated.
	
Example:
  ipam block available 10.0.0.0/16
  ipam block available 10.0.0.0/16 --file prod`,
	Args: cobra.ExactArgs(1),
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
		if err := ipam.PrintValidationResults(results); err != nil {
			fmt.Fprintln(os.Stderr, "Error printing validation results:", err)
			os.Exit(1)
		}

		if results.ErrorCount > 0 {
			os.Exit(1)
		}
	},
}

// blockUtilCommand represents the utilization command
var blockUtilCommand = &cobra.Command{
	Use:   "util [block-cidr]",
	Short: "Show IP address utilization",
	Long: `Show utilization statistics for a specific block or all blocks.

This command calculates IP address utilization statistics, showing:
- Total IP addresses in the block
- Allocated IP addresses (used by subnets)
- Available IP addresses
- Utilization percentage
- Subnet breakdown with allocation percentages

To show utilization for a specific block, provide its CIDR.
To show utilization for all blocks, omit the CIDR parameter.

Example:
  ipam block util 10.0.0.0/16   # Show utilization for a specific block
  ipam block util                # Show utilization for all blocks
`,
	Run: func(cmd *cobra.Command, args []string) {
		fileKey, _ := cmd.Flags().GetString("file")

		if len(args) > 0 {
			// Show utilization for a specific block
			cidr := args[0]
			err := ipam.PrintBlockUtilization(cfg, cidr, fileKey)
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error:", err)
				os.Exit(1)
			}
		} else {
			// Show utilization for all blocks
			err := ipam.PrintAllBlocksUtilization(cfg, fileKey)
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error:", err)
				os.Exit(1)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(blockCmd)
	blockCmd.AddCommand(blockCreateCmd)
	blockCmd.AddCommand(blockListCmd)
	blockCmd.AddCommand(blockShowCmd)
	blockCmd.AddCommand(blockDeleteCmd)
	blockCmd.AddCommand(blockAvailableCmd)
	blockCmd.AddCommand(blockValidateCmd)
	blockCmd.AddCommand(blockUtilCommand)

	blockCreateCmd.Flags().String("cidr", "", "CIDR range of the block")
	blockCreateCmd.Flags().String("description", "", "Description of the block")
	blockCreateCmd.Flags().StringP("file", "f", "default", "Block file key to use")
	blockCreateCmd.MarkFlagRequired("cidr")

	blockListCmd.Flags().StringP("file", "f", "default", "Block file key to use")

	blockShowCmd.Flags().StringP("file", "f", "default", "Block file key to use")

	blockDeleteCmd.Flags().Bool("force", false, "Force deletion of the block if it contains subnets")
	blockDeleteCmd.Flags().StringP("file", "f", "default", "Block file key to use")

	blockAvailableCmd.Flags().StringP("file", "f", "default", "Block file key to use")

	blockUtilCommand.Flags().StringP("file", "f", "default", "Block file key to use")
}