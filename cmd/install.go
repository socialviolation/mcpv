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
	Use:     "install [server@version]",
	Aliases: []string{"i"},
	Short:   "Install MCP servers",
	Long: `Install MCP servers either from mcpv.json configuration file or by specifying a server directly.

By default, servers are installed and configured for all detected AI agents. Use the --agent flag
to install and configure a server for a specific agent only.

Examples:
  mcpv install                              # Install all servers from mcpv.json for all agents
  mcpv install --agent roocode              # Install all servers from mcpv.json for RooCode only
  mcpv install server@1.0.0                 # Install specific server version for all agents
  mcpv install server --agent claude        # Install latest version for Claude Desktop only
  mcpv install server --repo <url> --agent cursor  # Install from repo for Cursor only

Use 'mcpv agents' to see supported agent types.`,
	RunE: runInstall,
}

func runInstall(cmd *cobra.Command, args []string) error {
	mgr, err := manager.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create manager: %w", err)
	}

	// Get agent flag
	agentFlag := cmd.Flag("agent").Value.String()
	var targetAgent manager.AgentType
	var agentSpecified bool

	if agentFlag != "" {
		agentSpecified = true

		// Get available agent types from the registry
		registry := mgr.GetAgentConfigManager().GetRegistry()
		availableTypes := registry.ListAgentTypes()

		// Check if the specified agent type is valid
		found := false
		for _, agentType := range availableTypes {
			if agentType == agentFlag {
				targetAgent = manager.AgentType(agentFlag)
				found = true
				break
			}
		}

		if !found {
			return fmt.Errorf("unsupported agent type: %s. Supported types: %v", agentFlag, availableTypes)
		}
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

		// Get global flag
		useGlobal, _ := cmd.Flags().GetBool("global")
		useLocal := !useGlobal

		if agentSpecified {
			return mgr.InstallFromConfigForAgentWithLocal(configPath, targetAgent, useLocal)
		}

		// Load config to check for default agent
		config, err := mgr.LoadProjectConfig(configPath)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// If there's a default agent, use it; otherwise install for all agents
		if config.DefaultAgent != "" {
			configType := "local"
			if useGlobal {
				configType = "global"
			}
			fmt.Printf("Using default agent: %s (%s config)\n", config.DefaultAgent, configType)
			return mgr.InstallFromConfigForAgentWithLocal(configPath, manager.AgentType(config.DefaultAgent), useLocal)
		}

		return mgr.InstallFromConfig(configPath)
	}

	// Install specific server
	repoURL := cmd.Flag("repo").Value.String()
	configPath := cmd.Flag("config").Value.String()
	if configPath == "" {
		configPath = "mcpv.json"
	}

	// Get global flag
	useGlobal, _ := cmd.Flags().GetBool("global")
	useLocal := !useGlobal

	// Load config to check for default agent if no agent specified
	var defaultAgentType manager.AgentType
	if !agentSpecified {
		config, err := mgr.LoadProjectConfig(configPath)
		if err == nil && config.DefaultAgent != "" {
			defaultAgentType = manager.AgentType(config.DefaultAgent)
			configType := "local"
			if useGlobal {
				configType = "global"
			}
			fmt.Printf("Using default agent: %s (%s config)\n", config.DefaultAgent, configType)
		}
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

		// Determine which agent to use
		var effectiveAgent manager.AgentType
		var effectiveAgentSpecified bool
		configType := "local"
		if useGlobal {
			configType = "global"
		}

		if agentSpecified {
			effectiveAgent = targetAgent
			effectiveAgentSpecified = true
			fmt.Printf("Installing %s@%s from %s for %s agent (%s config)...\n", name, version, repoURL, agentFlag, configType)
		} else if defaultAgentType != "" {
			effectiveAgent = defaultAgentType
			effectiveAgentSpecified = true
			fmt.Printf("Installing %s@%s from %s for default agent %s (%s config)...\n", name, version, repoURL, defaultAgentType, configType)
		} else {
			fmt.Printf("Installing %s@%s from %s...\n", name, version, repoURL)
		}

		// Install the server and add to config
		if effectiveAgentSpecified {
			err := mgr.InstallServerAndAddToConfigForAgentWithLocal(name, version, repoURL, configPath, effectiveAgent, useLocal)
			if err != nil {
				if strings.Contains(err.Error(), "already installed") {
					fmt.Printf("Server %s@%s is already installed\n", name, version)
					continue
				}
				return fmt.Errorf("failed to install server %s@%s: %w", name, version, err)
			}
			fmt.Printf("Successfully installed %s@%s for %s agent (%s config) and added to %s\n", name, version, effectiveAgent, configType, configPath)
		} else {
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
	}

	return nil
}

func init() {
	rootCmd.AddCommand(installCmd)
	installCmd.Flags().StringP("config", "c", "", "Path to mcpv.json config file")
	installCmd.Flags().StringP("repo", "r", "", "Repository URL for the server")
	installCmd.Flags().StringP("agent", "a", "", "Install server for specific agent only. If not specified, uses default agent from config. Use 'mcpv agents' to see available types")
	installCmd.Flags().BoolP("global", "g", false, "Install to global agent configuration instead of local (project-specific)")
}
