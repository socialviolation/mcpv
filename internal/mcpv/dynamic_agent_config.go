package manager

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// DynamicAgentConfig implements AgentConfig using the agent registry
type DynamicAgentConfig struct {
	agentType string
	registry  *AgentRegistry
}

// NewDynamicAgentConfig creates a new dynamic agent configuration
func NewDynamicAgentConfig(agentType string, registry *AgentRegistry) (*DynamicAgentConfig, error) {
	if _, exists := registry.GetAgent(agentType); !exists {
		return nil, fmt.Errorf("unknown agent type: %s", agentType)
	}

	return &DynamicAgentConfig{
		agentType: agentType,
		registry:  registry,
	}, nil
}

// NewDynamicAgentConfigWithLocal creates a new dynamic agent configuration with local preference
// This is kept for backward compatibility but now ignores the useLocal parameter
func NewDynamicAgentConfigWithLocal(agentType string, registry *AgentRegistry, useLocal bool) (*DynamicAgentConfig, error) {
	return NewDynamicAgentConfig(agentType, registry)
}

// GetConfigPath returns the path to the agent's configuration file
func (d *DynamicAgentConfig) GetConfigPath() (string, error) {
	spec, exists := d.registry.GetAgent(d.agentType)
	if !exists {
		return "", fmt.Errorf("unknown agent type: %s", d.agentType)
	}

	// For global agents, use global config path, otherwise use local
	useLocal := !spec.Global
	return d.registry.GetConfigPath(d.agentType, useLocal)
}

// GetConfigPathWithLocal returns the path to the agent's configuration file with local preference
func (d *DynamicAgentConfig) GetConfigPathWithLocal(useLocal bool) (string, error) {
	return d.registry.GetConfigPath(d.agentType, useLocal)
}

// LoadConfig loads the agent's configuration
func (d *DynamicAgentConfig) LoadConfig() (map[string]interface{}, error) {
	configPath, err := d.GetConfigPath()
	if err != nil {
		return nil, err
	}

	// If config doesn't exist, return empty structure based on agent spec
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		spec, _ := d.registry.GetAgent(d.agentType)
		if spec.Config != nil && spec.Config.Structure != nil {
			// Return a copy of the default structure
			result := make(map[string]interface{})
			for k, v := range spec.Config.Structure {
				result[k] = v
			}
			return result, nil
		}
		return map[string]interface{}{
			"mcpServers": map[string]interface{}{},
		}, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config map[string]interface{}
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Ensure mcpServers key exists
	if _, exists := config["mcpServers"]; !exists {
		config["mcpServers"] = map[string]interface{}{}
	}

	return config, nil
}

// SaveConfig saves the agent's configuration
func (d *DynamicAgentConfig) SaveConfig(config map[string]interface{}) error {
	configPath, err := d.GetConfigPath()
	if err != nil {
		return err
	}

	// Ensure directory exists
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// AddMCPServer adds an MCP server to the agent's configuration
func (d *DynamicAgentConfig) AddMCPServer(server *MCPServer) error {
	config, err := d.LoadConfig()
	if err != nil {
		return err
	}

	mcpServers, ok := config["mcpServers"].(map[string]interface{})
	if !ok {
		mcpServers = make(map[string]interface{})
		config["mcpServers"] = mcpServers
	}

	// Create server configuration based on agent's format specification
	spec, _ := d.registry.GetAgent(d.agentType)
	serverConfig := d.createServerConfig(server, spec)

	mcpServers[server.Name] = serverConfig

	return d.SaveConfig(config)
}

// RemoveMCPServer removes an MCP server from the agent's configuration
func (d *DynamicAgentConfig) RemoveMCPServer(serverName string) error {
	config, err := d.LoadConfig()
	if err != nil {
		return err
	}

	mcpServers, ok := config["mcpServers"].(map[string]interface{})
	if !ok {
		return nil // Nothing to remove
	}

	delete(mcpServers, serverName)

	return d.SaveConfig(config)
}

// createServerConfig creates a server configuration based on the agent's format specification
func (d *DynamicAgentConfig) createServerConfig(server *MCPServer, spec *AgentSpec) map[string]interface{} {
	serverConfig := make(map[string]interface{})

	// Add command
	if server.Command != "" {
		serverConfig["command"] = server.Command
	}

	// Add args
	if len(server.Args) > 0 {
		serverConfig["args"] = server.Args
	}

	// Add environment variables if they exist
	if len(server.Env) > 0 {
		serverConfig["env"] = server.Env
	}

	return serverConfig
}

// DynamicAgentConfigManager manages configurations for different AI agents using the registry
type DynamicAgentConfigManager struct {
	registry *AgentRegistry
	configs  map[string]AgentConfig
}

// NewDynamicAgentConfigManager creates a new dynamic agent configuration manager
func NewDynamicAgentConfigManager() (*DynamicAgentConfigManager, error) {
	registry, err := LoadAgentRegistry()
	if err != nil {
		return nil, fmt.Errorf("failed to load agent registry: %w", err)
	}

	configs := make(map[string]AgentConfig)

	// Initialize configs for all available agents
	for agentType := range registry.Agents {
		config, err := NewDynamicAgentConfig(agentType, registry)
		if err != nil {
			continue // Skip agents that can't be initialized
		}
		configs[agentType] = config
	}

	return &DynamicAgentConfigManager{
		registry: registry,
		configs:  configs,
	}, nil
}

// DetectAvailableAgents detects which AI agents are available on the system
func (dacm *DynamicAgentConfigManager) DetectAvailableAgents() []AgentType {
	availableTypes := dacm.registry.DetectAvailableAgents()

	// Convert strings to AgentType
	var result []AgentType
	for _, agentType := range availableTypes {
		result = append(result, AgentType(agentType))
	}

	return result
}

// AddServerToAgent adds an MCP server to a specific agent's configuration
func (dacm *DynamicAgentConfigManager) AddServerToAgent(agentType AgentType, server *MCPServer) error {
	config, exists := dacm.configs[string(agentType)]
	if !exists {
		return fmt.Errorf("unsupported agent type: %s", agentType)
	}

	return config.AddMCPServer(server)
}

// AddServerToAgentWithLocal adds an MCP server to a specific agent's configuration with local preference
func (dacm *DynamicAgentConfigManager) AddServerToAgentWithLocal(agentType AgentType, server *MCPServer, useLocal bool) error {
	// Create a new config with the specified local preference
	config, err := NewDynamicAgentConfigWithLocal(string(agentType), dacm.registry, useLocal)
	if err != nil {
		return fmt.Errorf("failed to create agent config: %w", err)
	}

	return config.AddMCPServer(server)
}

// RemoveServerFromAgent removes an MCP server from a specific agent's configuration
func (dacm *DynamicAgentConfigManager) RemoveServerFromAgent(agentType AgentType, serverName string) error {
	config, exists := dacm.configs[string(agentType)]
	if !exists {
		return fmt.Errorf("unsupported agent type: %s", agentType)
	}

	return config.RemoveMCPServer(serverName)
}

// AddServerToAllAgents adds an MCP server to all available agents
func (dacm *DynamicAgentConfigManager) AddServerToAllAgents(server *MCPServer) error {
	availableAgents := dacm.DetectAvailableAgents()

	var errors []string
	for _, agentType := range availableAgents {
		if err := dacm.AddServerToAgent(agentType, server); err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", agentType, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to add server to some agents: %v", errors)
	}

	return nil
}

// RemoveServerFromAllAgents removes an MCP server from all available agents
func (dacm *DynamicAgentConfigManager) RemoveServerFromAllAgents(serverName string) error {
	availableAgents := dacm.DetectAvailableAgents()

	var errors []string
	for _, agentType := range availableAgents {
		if err := dacm.RemoveServerFromAgent(agentType, serverName); err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", agentType, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to remove server from some agents: %v", errors)
	}

	return nil
}

// GetRegistry returns the agent registry
func (dacm *DynamicAgentConfigManager) GetRegistry() *AgentRegistry {
	return dacm.registry
}
