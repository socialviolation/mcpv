package manager

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// AgentType represents different AI agent types
type AgentType string

const (
	AgentTypeRooCode    AgentType = "roocode"
	AgentTypeClaude     AgentType = "claude"
	AgentTypeCursor     AgentType = "cursor"
	AgentTypeAider      AgentType = "aider"
	AgentTypeClaudeCode AgentType = "claude_code"
	AgentTypeWindsurf   AgentType = "windsurf"
)

// AgentConfig represents the configuration structure for different agents
type AgentConfig interface {
	GetConfigPath() (string, error)
	LoadConfig() (map[string]interface{}, error)
	SaveConfig(config map[string]interface{}) error
	AddMCPServer(server *MCPServer) error
	RemoveMCPServer(serverName string) error
}

// RooCodeConfig handles .roo/mcp.json configuratio`n
type RooCodeConfig struct {
	homeDir string
}

// NewRooCodeConfig creates a new RooCode configuration handler
func NewRooCodeConfig() (*RooCodeConfig, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	return &RooCodeConfig{
		homeDir: homeDir,
	}, nil
}

// GetConfigPath returns the path to the RooCode MCP configuration file
func (r *RooCodeConfig) GetConfigPath() (string, error) {
	return filepath.Join(r.homeDir, ".roo", "mcp.json"), nil
}

// LoadConfig loads the RooCode MCP configuration
func (r *RooCodeConfig) LoadConfig() (map[string]interface{}, error) {
	configPath, err := r.GetConfigPath()
	if err != nil {
		return nil, err
	}

	// If config doesn't exist, return empty structure
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
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

// SaveConfig saves the RooCode MCP configuration
func (r *RooCodeConfig) SaveConfig(config map[string]interface{}) error {
	configPath, err := r.GetConfigPath()
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

// AddMCPServer adds an MCP server to the RooCode configuration
func (r *RooCodeConfig) AddMCPServer(server *MCPServer) error {
	config, err := r.LoadConfig()
	if err != nil {
		return err
	}

	mcpServers, ok := config["mcpServers"].(map[string]interface{})
	if !ok {
		mcpServers = make(map[string]interface{})
		config["mcpServers"] = mcpServers
	}

	// Create server configuration for RooCode format
	serverConfig := map[string]interface{}{
		"command": server.Command,
		"args":    server.Args,
	}

	// Add environment variables if they exist
	if len(server.Env) > 0 {
		serverConfig["env"] = server.Env
	}

	mcpServers[server.Name] = serverConfig

	return r.SaveConfig(config)
}

// RemoveMCPServer removes an MCP server from the RooCode configuration
func (r *RooCodeConfig) RemoveMCPServer(serverName string) error {
	config, err := r.LoadConfig()
	if err != nil {
		return err
	}

	mcpServers, ok := config["mcpServers"].(map[string]interface{})
	if !ok {
		return nil // Nothing to remove
	}

	delete(mcpServers, serverName)

	return r.SaveConfig(config)
}

// ClaudeConfig handles Claude Desktop configuration
type ClaudeConfig struct {
	homeDir string
}

// NewClaudeConfig creates a new Claude configuration handler
func NewClaudeConfig() (*ClaudeConfig, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	return &ClaudeConfig{
		homeDir: homeDir,
	}, nil
}

// GetConfigPath returns the path to the Claude Desktop configuration file
func (c *ClaudeConfig) GetConfigPath() (string, error) {
	// Claude Desktop config location varies by OS
	switch {
	case fileExists(filepath.Join(c.homeDir, "Library", "Application Support", "Claude", "claude_desktop_config.json")):
		return filepath.Join(c.homeDir, "Library", "Application Support", "Claude", "claude_desktop_config.json"), nil
	case fileExists(filepath.Join(c.homeDir, ".config", "Claude", "claude_desktop_config.json")):
		return filepath.Join(c.homeDir, ".config", "Claude", "claude_desktop_config.json"), nil
	default:
		// Default to macOS location
		return filepath.Join(c.homeDir, "Library", "Application Support", "Claude", "claude_desktop_config.json"), nil
	}
}

// LoadConfig loads the Claude Desktop configuration
func (c *ClaudeConfig) LoadConfig() (map[string]interface{}, error) {
	configPath, err := c.GetConfigPath()
	if err != nil {
		return nil, err
	}

	// If config doesn't exist, return empty structure
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
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

// SaveConfig saves the Claude Desktop configuration
func (c *ClaudeConfig) SaveConfig(config map[string]interface{}) error {
	configPath, err := c.GetConfigPath()
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

// AddMCPServer adds an MCP server to the Claude Desktop configuration
func (c *ClaudeConfig) AddMCPServer(server *MCPServer) error {
	config, err := c.LoadConfig()
	if err != nil {
		return err
	}

	mcpServers, ok := config["mcpServers"].(map[string]interface{})
	if !ok {
		mcpServers = make(map[string]interface{})
		config["mcpServers"] = mcpServers
	}

	// Create server configuration for Claude Desktop format
	serverConfig := map[string]interface{}{
		"command": server.Command,
		"args":    server.Args,
	}

	// Add environment variables if they exist
	if len(server.Env) > 0 {
		serverConfig["env"] = server.Env
	}

	mcpServers[server.Name] = serverConfig

	return c.SaveConfig(config)
}

// RemoveMCPServer removes an MCP server from the Claude Desktop configuration
func (c *ClaudeConfig) RemoveMCPServer(serverName string) error {
	config, err := c.LoadConfig()
	if err != nil {
		return err
	}

	mcpServers, ok := config["mcpServers"].(map[string]interface{})
	if !ok {
		return nil // Nothing to remove
	}

	delete(mcpServers, serverName)

	return c.SaveConfig(config)
}

// AgentConfigManager manages configurations for different AI agents
type AgentConfigManager struct {
	configs map[AgentType]AgentConfig
}

// NewAgentConfigManager creates a new agent configuration manager
func NewAgentConfigManager() (*AgentConfigManager, error) {
	configs := make(map[AgentType]AgentConfig)

	// Initialize RooCode config
	rooConfig, err := NewRooCodeConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize RooCode config: %w", err)
	}
	configs[AgentTypeRooCode] = rooConfig

	// Initialize Claude config
	claudeConfig, err := NewClaudeConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Claude config: %w", err)
	}
	configs[AgentTypeClaude] = claudeConfig

	return &AgentConfigManager{
		configs: configs,
	}, nil
}

// DetectAvailableAgents detects which AI agents are available on the system
func (acm *AgentConfigManager) DetectAvailableAgents() []AgentType {
	var available []AgentType

	for agentType, config := range acm.configs {
		configPath, err := config.GetConfigPath()
		if err != nil {
			continue
		}

		// Check if config directory exists or can be created
		configDir := filepath.Dir(configPath)
		if _, err := os.Stat(configDir); err == nil {
			available = append(available, agentType)
		} else if os.IsNotExist(err) {
			// Try to create the directory to see if it's possible
			if err := os.MkdirAll(configDir, 0755); err == nil {
				available = append(available, agentType)
				// Remove the directory we just created for testing
				os.Remove(configDir)
			}
		}
	}

	return available
}

// AddServerToAgent adds an MCP server to a specific agent's configuration
func (acm *AgentConfigManager) AddServerToAgent(agentType AgentType, server *MCPServer) error {
	config, exists := acm.configs[agentType]
	if !exists {
		return fmt.Errorf("unsupported agent type: %s", agentType)
	}

	return config.AddMCPServer(server)
}

// RemoveServerFromAgent removes an MCP server from a specific agent's configuration
func (acm *AgentConfigManager) RemoveServerFromAgent(agentType AgentType, serverName string) error {
	config, exists := acm.configs[agentType]
	if !exists {
		return fmt.Errorf("unsupported agent type: %s", agentType)
	}

	return config.RemoveMCPServer(serverName)
}

// AddServerToAllAgents adds an MCP server to all available agents
func (acm *AgentConfigManager) AddServerToAllAgents(server *MCPServer) error {
	availableAgents := acm.DetectAvailableAgents()

	var errors []string
	for _, agentType := range availableAgents {
		if err := acm.AddServerToAgent(agentType, server); err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", agentType, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to add server to some agents: %v", errors)
	}

	return nil
}

// RemoveServerFromAllAgents removes an MCP server from all available agents
func (acm *AgentConfigManager) RemoveServerFromAllAgents(serverName string) error {
	availableAgents := acm.DetectAvailableAgents()

	var errors []string
	for _, agentType := range availableAgents {
		if err := acm.RemoveServerFromAgent(agentType, serverName); err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", agentType, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to remove server from some agents: %v", errors)
	}

	return nil
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
