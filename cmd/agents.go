package cmd

import (
	"fmt"

	manager "github.com/socialviolation/mcpv/internal/mcpv"
	"github.com/spf13/cobra"
)

// agentsCmd represents the agents command
var agentsCmd = &cobra.Command{
	Use:   "agents",
	Short: "Manage AI agent configurations",
	Long: `Manage AI agent configurations for MCP servers.

This command allows you to:
- List detected AI agents and their configuration paths
- View which agents have MCP servers configured
- Manually add or remove servers from specific agents

Examples:
  mcpv agents list                    # List detected agents
  mcpv agents add server-name roocode # Add server to specific agent
  mcpv agents remove server-name      # Remove server from all agents`,
}

// agentsListCmd represents the agents list command
var agentsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List detected AI agents",
	Long:  `List all detected AI agents and their configuration file paths.`,
	RunE:  runAgentsList,
}

// agentsAddCmd represents the agents add command
var agentsAddCmd = &cobra.Command{
	Use:   "add [server-name] [agent-type]",
	Short: "Add MCP server to specific agent",
	Long: `Add an installed MCP server to a specific AI agent's configuration.

Use 'mcpv agents list' to see available agent types and their configuration paths.

Examples:
  mcpv agents add my-server roocode   # Add server to RooCode
  mcpv agents add my-server claude    # Add server to Claude Desktop`,
	Args: cobra.ExactArgs(2),
	RunE: runAgentsAdd,
}

// agentsRemoveCmd represents the agents remove command
var agentsRemoveCmd = &cobra.Command{
	Use:   "remove [server-name]",
	Short: "Remove MCP server from agent configurations",
	Long: `Remove an MCP server from all AI agent configurations.

This will remove the server from all detected agents where it's currently configured.

Examples:
  mcpv agents remove my-server        # Remove server from all agents`,
	Args: cobra.ExactArgs(1),
	RunE: runAgentsRemove,
}

func runAgentsList(cmd *cobra.Command, args []string) error {
	mgr, err := manager.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create manager: %w", err)
	}

	return mgr.ListAgentConfigurations()
}

func runAgentsAdd(cmd *cobra.Command, args []string) error {
	serverName := args[0]
	agentTypeStr := args[1]

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
		if agentType == agentTypeStr {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("unsupported agent type: %s. Supported types: %v", agentTypeStr, availableTypes)
	}

	agentType := manager.AgentType(agentTypeStr)

	// Find the installed server
	servers, err := mgr.ListInstalledServers()
	if err != nil {
		return fmt.Errorf("failed to list installed servers: %w", err)
	}

	var targetServer *manager.MCPServer
	for _, server := range servers {
		if server.Name == serverName {
			targetServer = &server
			break
		}
	}

	if targetServer == nil {
		return fmt.Errorf("server %s is not installed. Use 'mcpv list' to see installed servers", serverName)
	}

	// We need to get the full server details including command and args
	// This is a limitation - we should store this info when listing servers
	fmt.Printf("Warning: Adding server without execution details. You may need to manually configure the command and args in the agent config.\n")

	// Add server to specific agent
	if err := mgr.AddServerToAgent(agentType, targetServer); err != nil {
		return fmt.Errorf("failed to add server to agent: %w", err)
	}

	fmt.Printf("Successfully added %s to %s configuration\n", serverName, agentType)
	return nil
}

func runAgentsRemove(cmd *cobra.Command, args []string) error {
	serverName := args[0]

	mgr, err := manager.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create manager: %w", err)
	}

	if err := mgr.RemoveServerFromAgentConfigs(serverName); err != nil {
		return fmt.Errorf("failed to remove server from agent configurations: %w", err)
	}

	fmt.Printf("Successfully removed %s from agent configurations\n", serverName)
	return nil
}

func init() {
	rootCmd.AddCommand(agentsCmd)
	agentsCmd.AddCommand(agentsListCmd)
	agentsCmd.AddCommand(agentsAddCmd)
	agentsCmd.AddCommand(agentsRemoveCmd)
}
