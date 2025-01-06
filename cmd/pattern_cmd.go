package cmd

import (
	"fmt"
	"os"

	"github.com/lugnut42/openipam/internal/ipam"
	"github.com/spf13/cobra"
)

var patternCmd = &cobra.Command{
	Use:   "pattern",
	Short: "Manage subnet allocation patterns",
	Long:  `Create, list, show, update, and delete subnet allocation patterns.`,
}

var patternCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new pattern",
	Long:  `Create a new subnet allocation pattern.`,
	Run: func(cmd *cobra.Command, args []string) {
		name, _ := cmd.Flags().GetString("name")
		cidrSize, _ := cmd.Flags().GetInt("cidr-size")
		environment, _ := cmd.Flags().GetString("environment")
		region, _ := cmd.Flags().GetString("region")
		block, _ := cmd.Flags().GetString("block")
		fileKey, _ := cmd.Flags().GetString("file")

		err := ipam.CreatePattern(cfg, name, cidrSize, environment, region, block, fileKey)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}

		fmt.Println("Pattern created successfully!")
	},
}

var patternListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available patterns",
	Long:  `List all available subnet allocation patterns.`,
	Run: func(cmd *cobra.Command, args []string) {
		fileKey, _ := cmd.Flags().GetString("file")

		err := ipam.ListPatterns(cfg, fileKey)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
	},
}

var patternShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show pattern details",
	Long:  `Show details of a specific subnet allocation pattern.`,
	Run: func(cmd *cobra.Command, args []string) {
		name, _ := cmd.Flags().GetString("name")
		fileKey, _ := cmd.Flags().GetString("file")

		err := ipam.ShowPattern(cfg, name, fileKey)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
	},
}

var patternDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a pattern",
	Long:  `Delete a specific subnet allocation pattern.`,
	Run: func(cmd *cobra.Command, args []string) {
		name, _ := cmd.Flags().GetString("name")
		fileKey, _ := cmd.Flags().GetString("file")

		err := ipam.DeletePattern(cfg, name, fileKey)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}

		fmt.Println("Pattern deleted successfully!")
	},
}

func init() {
	rootCmd.AddCommand(patternCmd)
	patternCmd.AddCommand(patternCreateCmd)
	patternCmd.AddCommand(patternListCmd)
	patternCmd.AddCommand(patternShowCmd)
	patternCmd.AddCommand(patternDeleteCmd)

	patternCreateCmd.Flags().StringP("name", "n", "", "Pattern name (required)")
	patternCreateCmd.MarkFlagRequired("name")
	patternCreateCmd.Flags().IntP("cidr-size", "c", 0, "CIDR size (required)")
	patternCreateCmd.MarkFlagRequired("cidr-size")
	patternCreateCmd.Flags().StringP("environment", "e", "", "Environment (required)")
	patternCreateCmd.MarkFlagRequired("environment")
	patternCreateCmd.Flags().StringP("region", "r", "", "Region (required)")
	patternCreateCmd.MarkFlagRequired("region")
	patternCreateCmd.Flags().StringP("block", "b", "", "Block CIDR (required)")
	patternCreateCmd.MarkFlagRequired("block")
	patternCreateCmd.Flags().StringP("file", "f", "default", "Key for the block file in the configuration (default is 'default')")

	patternListCmd.Flags().StringP("file", "f", "default", "Key for the block file in the configuration (default is 'default')")

	patternShowCmd.Flags().StringP("name", "n", "", "Pattern name (required)")
	patternShowCmd.MarkFlagRequired("name")
	patternShowCmd.Flags().StringP("file", "f", "default", "Key for the block file in the configuration (default is 'default')")

	patternDeleteCmd.Flags().StringP("name", "n", "", "Pattern name (required)")
	patternDeleteCmd.MarkFlagRequired("name")
	patternDeleteCmd.Flags().StringP("file", "f", "default", "Key for the block file in the configuration (default is 'default')")
}