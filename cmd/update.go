package cmd

import (
	"fmt"

	manager "github.com/socialviolation/mcpv/internal/mcpv"
	"github.com/spf13/cobra"
)

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update [server]",
	Short: "Update MCP servers to latest versions",
	Long: `Update MCP servers to their latest versions. If no server is specified,
updates all servers configured in mcpv.json.`,
	RunE: runUpdate,
}

func runUpdate(cmd *cobra.Command, args []string) error {
	mgr, err := manager.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create manager: %w", err)
	}

	// If no arguments provided, update all servers from mcpv.json
	if len(args) == 0 {
		return updateFromConfig(mgr, cmd)
	}

	// Update specific servers
	for _, serverName := range args {
		if err := updateServer(mgr, serverName); err != nil {
			return err
		}
	}

	return nil
}

func updateFromConfig(mgr *manager.Manager, cmd *cobra.Command) error {
	configPath := cmd.Flag("config").Value.String()
	if configPath == "" {
		configPath = "mcpv.json"
	}

	config, err := mgr.LoadProjectConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load project config: %w", err)
	}

	if len(config.Servers) == 0 {
		fmt.Println("No servers configured in mcpv.json")
		return nil
	}

	for _, server := range config.Servers {
		if server.Repository == "" {
			fmt.Printf("Skipping %s: no repository specified\n", server.Name)
			continue
		}

		fmt.Printf("Updating %s...\n", server.Name)
		if err := mgr.UpdateServer(server.Name, server.Repository); err != nil {
			fmt.Printf("Failed to update %s: %v\n", server.Name, err)
			continue
		}
		fmt.Printf("Successfully updated %s\n", server.Name)
	}

	return nil
}

func updateServer(mgr *manager.Manager, serverName string) error {
	// For now, we need the repository URL which we don't have
	// In a real implementation, you might store this information
	// or have a registry of known servers
	return fmt.Errorf("updating specific servers requires repository information. Use 'mcpv update' to update all servers from mcpv.json")
}

func init() {
	rootCmd.AddCommand(updateCmd)
	updateCmd.Flags().StringP("config", "c", "", "Path to mcpv.json config file")
}
