package manager

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"gopkg.in/yaml.v3"
)

// AgentRegistry holds the configuration for all supported agents
type AgentRegistry struct {
	Version         string                `yaml:"version" json:"version"`
	Agents          map[string]*AgentSpec `yaml:"agents" json:"agents"`
	ConfigDirectory *ConfigDirectorySpec  `yaml:"config_directory" json:"config_directory"`
	configPath      string                // Path to the agents.json file
}

// AgentSpec defines the configuration specification for an agent
type AgentSpec struct {
	Name               string            `yaml:"name" json:"name"`
	Type               string            `yaml:"type" json:"type"`
	Description        string            `yaml:"description" json:"description"`
	GlobalConfig       *ConfigSpec       `yaml:"global_config" json:"global_config"`
	LocalConfig        *ConfigSpec       `yaml:"local_config" json:"local_config"`
	ServerConfigFormat map[string]string `yaml:"server_config_format" json:"server_config_format"`
	Detection          *DetectionSpec    `yaml:"detection" json:"detection"`
}

// ConfigSpec defines how to handle configuration files
type ConfigSpec struct {
	Path      string                 `yaml:"path,omitempty" json:"path,omitempty"`
	Paths     []string               `yaml:"paths,omitempty" json:"paths,omitempty"`
	Format    string                 `yaml:"format" json:"format"`
	Structure map[string]interface{} `yaml:"structure" json:"structure"`
}

// DetectionSpec defines how to detect if an agent is available
type DetectionSpec struct {
	Paths    []string `yaml:"paths" json:"paths"`
	Commands []string `yaml:"commands" json:"commands"`
}

// ConfigDirectorySpec defines where to store mcpv configuration
type ConfigDirectorySpec struct {
	Name       string            `yaml:"name" json:"name"`
	Paths      map[string]string `yaml:"paths" json:"paths"`
	AgentsFile string            `yaml:"agents_file" json:"agents_file"`
}

// LoadAgentRegistry loads the agent registry from embedded YAML or agents.json
func LoadAgentRegistry() (*AgentRegistry, error) {
	// First try to load from platform-specific config directory
	configDir, err := getConfigDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get config directory: %w", err)
	}

	agentsPath := filepath.Join(configDir, "agents.json")

	// Try to load existing agents.json
	if _, err := os.Stat(agentsPath); err == nil {
		registry, err := loadAgentRegistryFromFile(agentsPath)
		if err == nil {
			registry.configPath = agentsPath
			return registry, nil
		}
		// If loading fails, fall back to embedded config
	}

	// Load from embedded YAML and install to config directory
	registry, err := loadEmbeddedAgentRegistry()
	if err != nil {
		return nil, fmt.Errorf("failed to load embedded agent registry: %w", err)
	}

	// Install to config directory
	if err := registry.InstallToConfigDir(); err != nil {
		return nil, fmt.Errorf("failed to install agent registry: %w", err)
	}

	registry.configPath = agentsPath
	return registry, nil
}

// loadAgentRegistryFromFile loads the registry from a JSON file
func loadAgentRegistryFromFile(path string) (*AgentRegistry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read agents file: %w", err)
	}

	var registry AgentRegistry
	if err := json.Unmarshal(data, &registry); err != nil {
		return nil, fmt.Errorf("failed to parse agents file: %w", err)
	}

	return &registry, nil
}

// loadEmbeddedAgentRegistry loads the registry from embedded YAML
func loadEmbeddedAgentRegistry() (*AgentRegistry, error) {
	// This will be replaced with embedded content at compile time
	yamlContent := getEmbeddedAgentsYAML()

	var registry AgentRegistry
	if err := yaml.Unmarshal([]byte(yamlContent), &registry); err != nil {
		return nil, fmt.Errorf("failed to parse embedded agents YAML: %w", err)
	}

	return &registry, nil
}

// InstallToConfigDir installs the agent registry to the platform-specific config directory
func (ar *AgentRegistry) InstallToConfigDir() error {
	configDir, err := getConfigDir()
	if err != nil {
		return fmt.Errorf("failed to get config directory: %w", err)
	}

	// Ensure config directory exists
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	agentsPath := filepath.Join(configDir, "agents.json")

	// Check if we need to update (version comparison)
	if shouldUpdate, err := ar.shouldUpdateConfig(agentsPath); err != nil {
		return fmt.Errorf("failed to check if update needed: %w", err)
	} else if !shouldUpdate {
		return nil // No update needed
	}

	// Write agents.json
	data, err := json.MarshalIndent(ar, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal agent registry: %w", err)
	}

	if err := os.WriteFile(agentsPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write agents.json: %w", err)
	}

	ar.configPath = agentsPath
	return nil
}

// shouldUpdateConfig determines if the config should be updated based on version
func (ar *AgentRegistry) shouldUpdateConfig(agentsPath string) (bool, error) {
	// If file doesn't exist, we need to create it
	if _, err := os.Stat(agentsPath); os.IsNotExist(err) {
		return true, nil
	}

	// Load existing config to compare versions
	existing, err := loadAgentRegistryFromFile(agentsPath)
	if err != nil {
		// If we can't load existing, assume we need to update
		return true, nil
	}

	// Compare versions (simple string comparison for now)
	// In a more sophisticated implementation, you'd use semantic versioning
	return ar.Version != existing.Version, nil
}

// getConfigDir returns the platform-specific configuration directory
func getConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(homeDir, "Library", "Application Support", "mcpv"), nil
	case "linux":
		if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
			return filepath.Join(xdgConfig, "mcpv"), nil
		}
		return filepath.Join(homeDir, ".config", "mcpv"), nil
	case "windows":
		if appData := os.Getenv("APPDATA"); appData != "" {
			return filepath.Join(appData, "mcpv"), nil
		}
		return filepath.Join(homeDir, "AppData", "Roaming", "mcpv"), nil
	default:
		return filepath.Join(homeDir, ".config", "mcpv"), nil
	}
}

// GetAgent returns an agent specification by type
func (ar *AgentRegistry) GetAgent(agentType string) (*AgentSpec, bool) {
	agent, exists := ar.Agents[agentType]
	return agent, exists
}

// ListAgentTypes returns all available agent types
func (ar *AgentRegistry) ListAgentTypes() []string {
	types := make([]string, 0, len(ar.Agents))
	for agentType := range ar.Agents {
		types = append(types, agentType)
	}
	return types
}

// DetectAvailableAgents detects which agents are available on the system
func (ar *AgentRegistry) DetectAvailableAgents() []string {
	var available []string

	for agentType, spec := range ar.Agents {
		if ar.isAgentAvailable(spec) {
			available = append(available, agentType)
		}
	}

	return available
}

// isAgentAvailable checks if an agent is available on the system
func (ar *AgentRegistry) isAgentAvailable(spec *AgentSpec) bool {
	// Check if any of the detection paths exist
	for _, path := range spec.Detection.Paths {
		expandedPath := expandPath(path)
		if _, err := os.Stat(expandedPath); err == nil {
			return true
		}
	}

	// Check if any of the detection commands are available
	for _, command := range spec.Detection.Commands {
		if _, err := exec.LookPath(command); err == nil {
			return true
		}
	}

	return false
}

// expandPath expands ~ and environment variables in paths
func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(homeDir, path[2:])
	}

	if strings.Contains(path, "%") && runtime.GOOS == "windows" {
		return os.ExpandEnv(path)
	}

	return path
}

// GetConfigPath returns the appropriate config path for an agent
func (ar *AgentRegistry) GetConfigPath(agentType string, useLocal bool) (string, error) {
	spec, exists := ar.GetAgent(agentType)
	if !exists {
		return "", fmt.Errorf("unknown agent type: %s", agentType)
	}

	var configSpec *ConfigSpec
	if useLocal {
		configSpec = spec.LocalConfig
	} else {
		configSpec = spec.GlobalConfig
	}

	if configSpec == nil {
		return "", fmt.Errorf("no config specification for agent %s", agentType)
	}

	// Handle single path
	if configSpec.Path != "" {
		return expandPath(configSpec.Path), nil
	}

	// Handle multiple paths (try each until one works)
	for _, path := range configSpec.Paths {
		expandedPath := expandPath(path)
		if _, err := os.Stat(filepath.Dir(expandedPath)); err == nil {
			return expandedPath, nil
		}
	}

	// If no existing paths found, return the first one
	if len(configSpec.Paths) > 0 {
		return expandPath(configSpec.Paths[0]), nil
	}

	return "", fmt.Errorf("no config path available for agent %s", agentType)
}

// Save saves the agent registry back to the config file
func (ar *AgentRegistry) Save() error {
	if ar.configPath == "" {
		return fmt.Errorf("no config path set")
	}

	data, err := json.MarshalIndent(ar, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal agent registry: %w", err)
	}

	if err := os.WriteFile(ar.configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write agents.json: %w", err)
	}

	return nil
}

// AddCustomAgent adds a custom agent to the registry
func (ar *AgentRegistry) AddCustomAgent(agentType string, spec *AgentSpec) error {
	if ar.Agents == nil {
		ar.Agents = make(map[string]*AgentSpec)
	}

	ar.Agents[agentType] = spec
	return ar.Save()
}

// RemoveCustomAgent removes a custom agent from the registry
func (ar *AgentRegistry) RemoveCustomAgent(agentType string) error {
	if ar.Agents == nil {
		return fmt.Errorf("agent %s not found", agentType)
	}

	if _, exists := ar.Agents[agentType]; !exists {
		return fmt.Errorf("agent %s not found", agentType)
	}

	delete(ar.Agents, agentType)
	return ar.Save()
}
