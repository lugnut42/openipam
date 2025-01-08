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
	Long:  `Add, list, show, and delete IP blocks.`,
}

var blockListCmd = &cobra.Command{
	Use:   "list",
	Short: "List IP blocks",
	Long:  `List all available IP blocks.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := ipam.ListBlocks(cfg)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
	},
}

var blockAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new IP block",
	Long:  `Add a new IP block to the YAML file.`,
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

		err := ipam.DeleteBlock(cfg, cidr, force)
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

func init() {
	rootCmd.AddCommand(blockCmd)
	blockCmd.AddCommand(blockListCmd)
	blockAddCmd.Flags().StringP("cidr", "c", "", "CIDR block (required)")
	blockAddCmd.MarkFlagRequired("cidr")
	blockAddCmd.Flags().StringP("description", "d", "", "Description of the block")
	blockAddCmd.Flags().StringP("file", "f", "default", "Key for the block file in the configuration (default is 'default')")
	blockCmd.AddCommand(blockAddCmd)
	blockShowCmd.Flags().StringP("file", "f", "default", "Key for the block file in the configuration (default is 'default')")
	blockCmd.AddCommand(blockShowCmd)
	blockDeleteCmd.Flags().BoolP("force", "f", false, "Force deletion without confirmation")
	blockCmd.AddCommand(blockDeleteCmd)
	blockAvailableCmd.Flags().StringP("file", "f", "default", "Key for the block file in the configuration (default is 'default')")
	blockCmd.AddCommand(blockAvailableCmd)
}
