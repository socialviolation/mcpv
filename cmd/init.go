package cmd

import (
	"fmt"
	"os"

	manager "github.com/socialviolation/mcpv/internal/mcpv"
	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init --agent <agent-type>",
	Short: "Initialize a new mcpv.json configuration file",
	Long: `Initialize a new mcpv.json configuration file in the current directory.
This file will contain the MCP server dependencies for your project.

A default agent must be specified to ensure proper configuration management.

Examples:
  mcpv init --agent claude        # Initialize with Claude Desktop as default agent
  mcpv init --agent roocode       # Initialize with RooCode as default agent
  mcpv init --agent cursor        # Initialize with Cursor as default agent

Use 'mcpv agents list' to see available agent types.`,
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

	// Get and validate agent flag
	agentFlag := cmd.Flag("agent").Value.String()
	if agentFlag == "" {
		return fmt.Errorf("default agent is required. Use --agent flag to specify one. Use 'mcpv agents list' to see available types")
	}

	mgr, err := manager.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create manager: %w", err)
	}

	// Get available agent types from the registry
	registry := mgr.GetAgentConfigManager().GetRegistry()
	availableTypes := registry.ListAgentTypes()

	// Check if the specified agent type is valid
	found := false
	for _, agentType := range availableTypes {
		if agentType == agentFlag {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("unsupported agent type: %s. Supported types: %v", agentFlag, availableTypes)
	}

	config := &manager.ProjectConfig{
		Servers:      []manager.MCPServer{},
		DefaultAgent: agentFlag,
	}

	if err := mgr.SaveProjectConfig(config, configPath); err != nil {
		return fmt.Errorf("failed to create mcpv.json: %w", err)
	}

	fmt.Printf("Created mcpv.json with default agent: %s\n", agentFlag)

	return nil
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().BoolP("force", "f", false, "Overwrite existing mcpv.json")
	initCmd.Flags().StringP("agent", "a", "", "Default agent type (required). Use 'mcpv agents list' to see available types")
	initCmd.MarkFlagRequired("agent")
}
