package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	manager "github.com/socialviolation/mcpv/internal/mcpv"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed MCP servers",
	Long: `List all installed MCP servers with their versions and installation paths.

Examples:
  mcpv list                       # List all installed servers
  mcpv list --project             # List servers required by current project`,
	RunE: runList,
}

func runList(cmd *cobra.Command, args []string) error {
	mgr, err := manager.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create manager: %w", err)
	}

	showProject, _ := cmd.Flags().GetBool("project")
	showInstalled, _ := cmd.Flags().GetBool("installed")

	// If --project is explicitly set, show project servers
	if showProject {
		return listProjectServers(mgr, cmd)
	}

	// If --installed is explicitly set, show installed servers
	if showInstalled {
		return listInstalledServers(mgr)
	}

	// Default behavior: if mcpv.json exists, show project servers, otherwise show installed
	configPath := cmd.Flag("config").Value.String()
	if configPath == "" {
		configPath = findConfigFile()
	}

	if _, err := os.Stat(configPath); err == nil {
		// mcpv.json exists, show project servers by default
		fmt.Printf("Project servers (from %s):\n", configPath)
		return listProjectServers(mgr, cmd)
	}

	// No mcpv.json found, show installed servers
	fmt.Println("Installed servers:")
	return listInstalledServers(mgr)
}

func listInstalledServers(mgr *manager.Manager) error {
	servers, err := mgr.ListInstalledServers()
	if err != nil {
		return fmt.Errorf("failed to list servers: %w", err)
	}

	if len(servers) == 0 {
		fmt.Println("No MCP servers installed")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tVERSION\tPATH")
	fmt.Fprintln(w, "----\t-------\t----")

	for _, server := range servers {
		fmt.Fprintf(w, "%s\t%s\t%s\n", server.Name, server.Version, server.InstallPath)
	}

	return w.Flush()
}

func listProjectServers(mgr *manager.Manager, cmd *cobra.Command) error {
	configPath := cmd.Flag("config").Value.String()
	if configPath == "" {
		configPath = findConfigFile()
	}

	config, err := mgr.LoadProjectConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load project config: %w", err)
	}

	if len(config.Servers) == 0 {
		fmt.Println("No servers configured in project")
		return nil
	}

	// Get installed servers to check status
	installed, err := mgr.ListInstalledServers()
	if err != nil {
		return fmt.Errorf("failed to list installed servers: %w", err)
	}

	installedMap := make(map[string]bool)
	for _, server := range installed {
		key := fmt.Sprintf("%s@%s", server.Name, server.Version)
		installedMap[key] = true
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tVERSION\tREPOSITORY\tSTATUS\tPATH")
	fmt.Fprintln(w, "----\t-------\t----------\t------\t----")

	for _, server := range config.Servers {
		version := server.Version
		if version == "" {
			version = "latest"
		}

		status := "Not Installed"
		installPath := ""

		// Find the install path if installed
		for _, installedServer := range installed {
			if installedServer.Name == server.Name && installedServer.Version == version {
				status = "Installed"
				installPath = installedServer.InstallPath
				break
			}
		}

		// If not installed, show where it would be installed
		if installPath == "" {
			installPath = filepath.Join(mgr.GetDataDir(), server.Name, version)
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", server.Name, version, server.Repository, status, installPath)
	}

	return w.Flush()
}

// findConfigFile looks for mcpv.json in current directory, then XDG_CONFIG_HOME
func findConfigFile() string {
	// First check current directory
	if _, err := os.Stat("mcpv.json"); err == nil {
		return "mcpv.json"
	}

	// Then check XDG_CONFIG_HOME or ~/.config
	var configDir string
	if xdgConfigHome := os.Getenv("XDG_CONFIG_HOME"); xdgConfigHome != "" {
		configDir = filepath.Join(xdgConfigHome, "mcpv")
	} else {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "mcpv.json" // fallback to current directory
		}
		configDir = filepath.Join(homeDir, ".config", "mcpv")
	}

	configPath := filepath.Join(configDir, "mcpv.json")
	if _, err := os.Stat(configPath); err == nil {
		return configPath
	}

	// Default fallback
	return "mcpv.json"
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().BoolP("project", "p", false, "List servers required by current project")
	listCmd.Flags().BoolP("installed", "i", false, "List installed servers")
	listCmd.Flags().StringP("config", "c", "", "Path to mcpv.json config file")
}
