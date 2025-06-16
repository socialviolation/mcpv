package cmd

import (
	"fmt"
	"os"
	"strings"

	manager "github.com/socialviolation/mcpv/internal/mcpv"
	"github.com/spf13/cobra"
)

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:   "install [server@version]",
	Short: "Install MCP servers",
	Long: `Install MCP servers either from mcpv.json configuration file or by specifying a server directly.

Examples:
  mcpv install                    # Install all servers from mcpv.json
  mcpv install server@1.0.0       # Install specific server version
  mcpv install server             # Install latest version of server`,
	RunE: runInstall,
}

func runInstall(cmd *cobra.Command, args []string) error {
	mgr, err := manager.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create manager: %w", err)
	}

	// If no arguments provided, install from mcpv.json
	if len(args) == 0 {
		configPath := cmd.Flag("config").Value.String()
		if configPath == "" {
			configPath = "mcpv.json"
		}

		// Check if config file exists
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			return fmt.Errorf("no mcpv.json found in current directory. Use 'mcpv init' to create one or specify a server directly")
		}

		return mgr.InstallFromConfig(configPath)
	}

	// Install specific server
	repoURL := cmd.Flag("repo").Value.String()
	configPath := cmd.Flag("config").Value.String()
	if configPath == "" {
		configPath = "mcpv.json"
	}

	for _, arg := range args {
		name, version := manager.ParseServerSpec(arg)

		if version == "" {
			version = "latest"
		}

		if repoURL == "" {
			// For now, we'll need the repository URL to be provided
			// In a real implementation, you might have a registry of known servers
			fmt.Printf("Installing specific servers requires repository URL.\n")
			fmt.Printf("Usage: mcpv install %s@%s --repo <repository-url>\n", name, version)
			fmt.Printf("Or add the server to mcpv.json first with: mcpv init\n")
			return fmt.Errorf("repository URL required for server installation")
		}

		fmt.Printf("Installing %s@%s from %s...\n", name, version, repoURL)

		// Install the server and add to config
		err := mgr.InstallServerAndAddToConfig(name, version, repoURL, configPath)
		if err != nil {
			if strings.Contains(err.Error(), "already installed") {
				fmt.Printf("Server %s@%s is already installed\n", name, version)
				continue
			}
			return fmt.Errorf("failed to install server %s@%s: %w", name, version, err)
		}

		fmt.Printf("Successfully installed %s@%s and added to %s\n", name, version, configPath)
	}

	return nil
}

func init() {
	rootCmd.AddCommand(installCmd)
	installCmd.Flags().StringP("config", "c", "", "Path to mcpv.json config file")
	installCmd.Flags().StringP("repo", "r", "", "Repository URL for the server")
}
