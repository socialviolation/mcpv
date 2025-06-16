package cmd

import (
	"fmt"

	manager "github.com/socialviolation/mcpv/internal/mcpv"
	"github.com/spf13/cobra"
)

// removeCmd represents the remove command
var removeCmd = &cobra.Command{
	Use:   "remove <server@version>",
	Short: "Remove installed MCP servers",
	Long:  `Remove installed MCP servers by specifying the server name and version.`,
	Args:  cobra.MinimumNArgs(1),
	RunE:  runRemove,
}

func runRemove(cmd *cobra.Command, args []string) error {
	mgr, err := manager.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create manager: %w", err)
	}

	for _, arg := range args {
		name, version := manager.ParseServerSpec(arg)

		if version == "" {
			// Remove all versions of the server
			if err := removeAllVersions(mgr, name); err != nil {
				return err
			}
		} else {
			// Remove specific version
			fmt.Printf("Removing %s@%s...\n", name, version)
			if err := mgr.RemoveServer(name, version); err != nil {
				return fmt.Errorf("failed to remove %s@%s: %w", name, version, err)
			}
			fmt.Printf("Successfully removed %s@%s\n", name, version)
		}
	}

	return nil
}

func removeAllVersions(mgr *manager.Manager, serverName string) error {
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
		return fmt.Errorf("server %s is not installed", serverName)
	}

	// Remove each version
	for _, version := range versionsToRemove {
		fmt.Printf("Removing %s@%s...\n", serverName, version)
		if err := mgr.RemoveServer(serverName, version); err != nil {
			return fmt.Errorf("failed to remove %s@%s: %w", serverName, version, err)
		}
		fmt.Printf("Successfully removed %s@%s\n", serverName, version)
	}

	return nil
}

func init() {
	rootCmd.AddCommand(removeCmd)
}
