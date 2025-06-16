package cmd

import (
	"fmt"

	manager "github.com/socialviolation/mcpv/internal/mcpv"
	"github.com/spf13/cobra"
)

// removeCmd represents the remove command
var removeCmd = &cobra.Command{
	Use:     "remove <server@version>",
	Aliases: []string{"rm"},
	Short:   "Remove installed MCP servers",
	Long:    `Remove installed MCP servers by specifying the server name and version.`,
	Args:    cobra.MinimumNArgs(1),
	RunE:    runRemove,
}

func runRemove(cmd *cobra.Command, args []string) error {
	mgr, err := manager.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create manager: %w", err)
	}

	configPath := cmd.Flag("config").Value.String()
	if configPath == "" {
		configPath = "mcpv.json"
	}

	for _, arg := range args {
		name, version := manager.ParseServerSpec(arg)

		if version == "" {
			// Remove all versions of the server
			if err := removeAllVersions(mgr, name, configPath); err != nil {
				return err
			}
		} else {
			// Remove specific version
			fmt.Printf("Removing %s@%s...\n", name, version)
			if err := mgr.RemoveServer(name, version); err != nil {
				// Continue even if server removal fails - we still want to remove from config
				fmt.Printf("Warning: Failed to remove installed server %s@%s: %v\n", name, version, err)
			} else {
				fmt.Printf("Successfully removed %s@%s\n", name, version)
			}

			// Always remove from mcpv.json config regardless of installation status
			if err := removeFromConfig(mgr, name, version, configPath); err != nil {
				fmt.Printf("Warning: Failed to remove %s@%s from %s: %v\n", name, version, configPath, err)
			} else {
				fmt.Printf("✓ Removed %s@%s from %s\n", name, version, configPath)
			}
		}
	}

	return nil
}

func removeAllVersions(mgr *manager.Manager, serverName string, configPath string) error {
	// Get all installed servers
	servers, err := mgr.ListInstalledServers()
	if err != nil {
		return fmt.Errorf("failed to list installed servers: %w", err)
	}

	// Find all versions of the specified server
	var versionsToRemove []string
	for _, server := range servers {
		if server.Name == serverName {
			versionsToRemove = append(versionsToRemove, server.Version)
		}
	}

	if len(versionsToRemove) == 0 {
		fmt.Printf("Warning: server %s is not installed, but will still remove from config\n", serverName)
	}

	// Remove each version
	for _, version := range versionsToRemove {
		fmt.Printf("Removing %s@%s...\n", serverName, version)
		if err := mgr.RemoveServer(serverName, version); err != nil {
			fmt.Printf("Warning: Failed to remove installed server %s@%s: %v\n", serverName, version, err)
		} else {
			fmt.Printf("Successfully removed %s@%s\n", serverName, version)
		}

		// Always remove from mcpv.json config regardless of installation status
		if err := removeFromConfig(mgr, serverName, version, configPath); err != nil {
			fmt.Printf("Warning: Failed to remove %s@%s from %s: %v\n", serverName, version, configPath, err)
		} else {
			fmt.Printf("✓ Removed %s@%s from %s\n", serverName, version, configPath)
		}
	}

	// Also try to remove any entries from config that might not have been installed
	if err := removeAllVersionsFromConfig(mgr, serverName, configPath); err != nil {
		fmt.Printf("Warning: Failed to remove all versions of %s from %s: %v\n", serverName, configPath, err)
	}

	return nil
}

// removeFromConfig removes a specific server version from mcpv.json
func removeFromConfig(mgr *manager.Manager, serverName, version, configPath string) error {
	config, err := mgr.LoadProjectConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Filter out the server to remove
	var updatedServers []manager.MCPServer
	found := false
	for _, server := range config.Servers {
		if server.Name == serverName && server.Version == version {
			found = true
			continue // Skip this server
		}
		updatedServers = append(updatedServers, server)
	}

	if !found {
		return fmt.Errorf("server %s@%s not found in config", serverName, version)
	}

	config.Servers = updatedServers
	return mgr.SaveProjectConfig(config, configPath)
}

// removeAllVersionsFromConfig removes all versions of a server from mcpv.json
func removeAllVersionsFromConfig(mgr *manager.Manager, serverName, configPath string) error {
	config, err := mgr.LoadProjectConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Filter out all versions of the server
	var updatedServers []manager.MCPServer
	removedCount := 0
	for _, server := range config.Servers {
		if server.Name == serverName {
			removedCount++
			continue // Skip this server
		}
		updatedServers = append(updatedServers, server)
	}

	if removedCount > 0 {
		config.Servers = updatedServers
		if err := mgr.SaveProjectConfig(config, configPath); err != nil {
			return err
		}
		fmt.Printf("✓ Removed %d version(s) of %s from %s\n", removedCount, serverName, configPath)
	}

	return nil
}

func init() {
	rootCmd.AddCommand(removeCmd)
	removeCmd.Flags().StringP("config", "c", "", "Path to mcpv.json config file")
	removeCmd.Flags().StringP("agent", "a", "", "Remove server from specific agent only. If not specified, uses default agent from config")
	removeCmd.Flags().BoolP("global", "g", false, "Remove from global agent configuration instead of local (project-specific)")
}
