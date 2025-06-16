package cmd

import (
	"fmt"
	"os"

	manager "github.com/socialviolation/mcpv/internal/mcpv"
	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new mcpv.json configuration file",
	Long: `Initialize a new mcpv.json configuration file in the current directory.
This file will contain the MCP server dependencies for your project.

Examples:
  mcpv init                       # Create mcpv.json with example configuration`,
	RunE: runInit,
}

func runInit(cmd *cobra.Command, args []string) error {
	configPath := "mcpv.json"

	// Check if config file already exists
	if _, err := os.Stat(configPath); err == nil {
		overwrite, _ := cmd.Flags().GetBool("force")
		if !overwrite {
			return fmt.Errorf("mcpv.json already exists. Use --force to overwrite")
		}
	}

	mgr, err := manager.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create manager: %w", err)
	}

	config := &manager.ProjectConfig{
		Servers: []manager.MCPServer{},
	}

	if err := mgr.SaveProjectConfig(config, configPath); err != nil {
		return fmt.Errorf("failed to create mcpv.json: %w", err)
	}

	fmt.Println("Created empty mcpv.json")

	return nil
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().BoolP("force", "f", false, "Overwrite existing mcpv.json")
}
