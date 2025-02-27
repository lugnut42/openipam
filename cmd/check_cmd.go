package cmd

import (
	"fmt"
	"os"

	"github.com/lugnut42/openipam/internal/ipam"
	"github.com/spf13/cobra"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check and validate IPAM configurations",
	Long:  `Check and validate the integrity of block files and subnets in the IPAM configuration.`,
}

var checkBlocksCmd = &cobra.Command{
	Use:   "blocks [file-key]",
	Short: "Check block file integrity",
	Long: `Validate block files for integrity, proper structure, and correct references.

This performs a comprehensive validation on block files, checking:
- YAML structure and syntax
- CIDR format of blocks and subnets
- Containment and overlap of subnets
- Duplicate entries and references
- Required fields and metadata

If a specific file-key is provided, only checks that file. Otherwise, checks all configured files.`,
	Run: func(cmd *cobra.Command, args []string) {
		all, _ := cmd.Flags().GetBool("all")
		
		if all {
			fmt.Println("Checking all block files...")
			err := ipam.ValidateAllBlockFiles(cfg)
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error:", err)
				os.Exit(1)
			}
		} else if len(args) > 0 {
			fileKey := args[0]
			fmt.Printf("Checking block file '%s'...\n", fileKey)
			results, err := ipam.ValidateBlockFile(cfg, fileKey)
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error:", err)
				os.Exit(1)
			}
			ipam.PrintValidationResults(results)
			if results.ErrorCount > 0 {
				os.Exit(1)
			}
		} else {
			// Default to checking the default block file
			fmt.Println("Checking default block file...")
			results, err := ipam.ValidateBlockFile(cfg, "default")
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error:", err)
				os.Exit(1)
			}
			ipam.PrintValidationResults(results)
			if results.ErrorCount > 0 {
				os.Exit(1)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(checkCmd)
	checkCmd.AddCommand(checkBlocksCmd)
	
	checkBlocksCmd.Flags().BoolP("all", "a", false, "Check all block files")
}