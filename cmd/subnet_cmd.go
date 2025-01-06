package cmd

import (
	"fmt"

	"github.com/lugnut42/openipam/internal/ipam"
	"github.com/spf13/cobra"
)

var subnetCmd = &cobra.Command{
	Use:   "subnet",
	Short: "Manage subnets",
	Long:  `Create, list, show, and delete subnets.`,
}

var subnetCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new subnet",
	Long:  `Allocate a new subnet within an existing IP block.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		block, _ := cmd.Flags().GetString("block")
		cidr, _ := cmd.Flags().GetString("cidr")
		name, _ := cmd.Flags().GetString("name")
		region, _ := cmd.Flags().GetString("region")

		err := ipam.CreateSubnet(cfg, block, cidr, name, region)
		if err != nil {
			return fmt.Errorf("Error: %w", err)
		}

		fmt.Println("Subnet created successfully!")
		return nil
	},
}

var subnetCreateFromPatternCmd = &cobra.Command{
	Use:   "create-from-pattern",
	Short: "Create a new subnet from a pattern",
	Long:  `Allocate a new subnet within an existing IP block based on a predefined pattern.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		patternName, _ := cmd.Flags().GetString("pattern")
		fileKey, _ := cmd.Flags().GetString("file")

		err := ipam.CreateSubnetFromPattern(cfg, patternName, fileKey)
		if err != nil {
			return fmt.Errorf("Error: %w", err)
		}

		fmt.Println("Subnet created successfully!")
		return nil
	},
}

var subnetDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a subnet",
	Long:  `Delete a subnet from an existing IP block.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cidr, _ := cmd.Flags().GetString("cidr")
		force, _ := cmd.Flags().GetBool("force")

		err := ipam.DeleteSubnet(cfg, cidr, force)
		if err != nil {
			return fmt.Errorf("Error: %w", err)
		}

		fmt.Println("Subnet deleted successfully!")
		return nil
	},
}

var subnetListCmd = &cobra.Command{
	Use:   "list",
	Short: "List subnets",
	Long:  `List all subnets within an existing IP block.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		block, _ := cmd.Flags().GetString("block")
		region, _ := cmd.Flags().GetString("region")

		err := ipam.ListSubnets(cfg, block, region)
		if err != nil {
			return fmt.Errorf("Error: %w", err)
		}
		return nil
	},
}

var subnetShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show details of a subnet",
	Long:  `Show details of a specific subnet within an existing IP block.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cidr, _ := cmd.Flags().GetString("cidr")

		err := ipam.ShowSubnet(cfg, cidr)
		if err != nil {
			return fmt.Errorf("Error: %w", err)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(subnetCmd)
	subnetCmd.AddCommand(subnetCreateCmd)
	subnetCmd.AddCommand(subnetCreateFromPatternCmd)
	subnetCmd.AddCommand(subnetDeleteCmd)
	subnetCmd.AddCommand(subnetListCmd)
	subnetCmd.AddCommand(subnetShowCmd)

	subnetCreateCmd.Flags().StringP("block", "b", "", "Block CIDR (required)")
	subnetCreateCmd.MarkFlagRequired("block")
	subnetCreateCmd.Flags().StringP("cidr", "c", "", "Subnet CIDR (required)")
	subnetCreateCmd.MarkFlagRequired("cidr")
	subnetCreateCmd.Flags().StringP("name", "n", "", "Subnet name (required)")
	subnetCreateCmd.MarkFlagRequired("name")
	subnetCreateCmd.Flags().StringP("region", "r", "", "Region (required)")
	subnetCreateCmd.MarkFlagRequired("region")

	subnetCreateFromPatternCmd.Flags().StringP("pattern", "p", "", "Pattern name (required)")
	subnetCreateFromPatternCmd.MarkFlagRequired("pattern")
	subnetCreateFromPatternCmd.Flags().StringP("file", "f", "default", "Key for the block file in the configuration (default is 'default')")

	subnetDeleteCmd.Flags().StringP("cidr", "c", "", "Subnet CIDR (required)")
	subnetDeleteCmd.MarkFlagRequired("cidr")
	subnetDeleteCmd.Flags().BoolP("force", "f", false, "Force delete")

	subnetListCmd.Flags().StringP("block", "b", "", "Block CIDR")
	subnetListCmd.Flags().StringP("region", "r", "", "Region")

	subnetShowCmd.Flags().StringP("cidr", "c", "", "Subnet CIDR (required)")
	subnetShowCmd.MarkFlagRequired("cidr")
}
