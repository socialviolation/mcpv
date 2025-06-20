package manager

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

// MCPServer represents an MCP server configuration
type MCPServer struct {
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	Repository  string            `json:"repository"`
	InstallPath string            `json:"install_path,omitempty"`
	Installed   bool              `json:"installed,omitempty"`
	Command     string            `json:"command,omitempty"`
	Args        []string          `json:"args,omitempty"`
	Env         map[string]string `json:"env,omitempty"`
}

// ProjectConfig represents the mcpv.json configuration file
type ProjectConfig struct {
	Servers      []MCPServer           `json:"servers"`
	DefaultAgent string                `json:"default_agent,omitempty"`
	Agents       map[string]*AgentSpec `json:"agents,omitempty"`
}

// Manager handles MCP server operations
type Manager struct {
	dataDir            string
	agentConfigManager *DynamicAgentConfigManager
}

// NewManager creates a new manager instance
func NewManager() (*Manager, error) {
	dataDir, err := getDataDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get data directory: %w", err)
	}

	// Ensure data directory exists
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	// Initialize dynamic agent config manager
	agentConfigManager, err := NewDynamicAgentConfigManager()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize agent config manager: %w", err)
	}

	return &Manager{
		dataDir:            dataDir,
		agentConfigManager: agentConfigManager,
	}, nil
}

// getDataDir returns the XDG_DATA_HOME directory or default
func getDataDir() (string, error) {
	if xdgDataHome := os.Getenv("XDG_DATA_HOME"); xdgDataHome != "" {
		return filepath.Join(xdgDataHome, "mcpv"), nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(homeDir, ".local", "share", "mcpv"), nil
}

// GetDataDir returns the data directory for this manager instance
func (m *Manager) GetDataDir() string {
	return m.dataDir
}

// LoadProjectConfig loads the mcpv.json configuration file
func (m *Manager) LoadProjectConfig(configPath string) (*ProjectConfig, error) {
	if configPath == "" {
		configPath = "mcpv.json"
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &ProjectConfig{Servers: []MCPServer{}}, nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config ProjectConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// SaveProjectConfig saves the mcpv.json configuration file
func (m *Manager) SaveProjectConfig(config *ProjectConfig, configPath string) error {
	if configPath == "" {
		configPath = "mcpv.json"
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

// InstallServer installs a specific MCP server version
func (m *Manager) InstallServer(name, version, repository string) (*MCPServer, error) {
	serverDir := filepath.Join(m.dataDir, name, version)

	// Check if already installed
	if _, err := os.Stat(serverDir); err == nil {
		return nil, fmt.Errorf("server %s@%s is already installed", name, version)
	}

	// Create server directory
	if err := os.MkdirAll(serverDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create server directory: %w", err)
	}

	// Clone the repository
	if err := m.cloneRepository(repository, version, serverDir); err != nil {
		// Clean up on failure
		os.RemoveAll(serverDir)
		return nil, fmt.Errorf("failed to clone repository: %w", err)
	}

	// Install dependencies if needed
	if err := m.installDependencies(serverDir); err != nil {
		return nil, fmt.Errorf("failed to install dependencies: %w", err)
	}

	// Build the server
	if err := m.buildServer(serverDir); err != nil {
		return nil, fmt.Errorf("failed to build server: %w", err)
	}

	// Determine execution configuration
	command, args, env, err := m.determineExecution(serverDir)
	if err != nil {
		return nil, fmt.Errorf("failed to determine execution configuration: %w", err)
	}

	server := &MCPServer{
		Name:        name,
		Version:     version,
		Repository:  repository,
		InstallPath: serverDir,
		Installed:   true,
		Command:     command,
		Args:        args,
		Env:         env,
	}

	return server, nil
}

// cloneRepository clones a git repository at a specific version
func (m *Manager) cloneRepository(repoURL, version, targetDir string) error {
	// Clone the repository
	repo, err := git.PlainClone(targetDir, false, &git.CloneOptions{
		URL: repoURL,
	})
	if err != nil {
		return err
	}

	// If version is specified, checkout that version
	if version != "" && version != "latest" {
		worktree, err := repo.Worktree()
		if err != nil {
			return err
		}

		// Try to checkout as tag first, then as branch
		err = worktree.Checkout(&git.CheckoutOptions{
			Branch: plumbing.ReferenceName("refs/tags/" + version),
		})
		if err != nil {
			// Try as branch
			err = worktree.Checkout(&git.CheckoutOptions{
				Branch: plumbing.ReferenceName("refs/heads/" + version),
			})
			if err != nil {
				return fmt.Errorf("failed to checkout version %s: %w", version, err)
			}
		}
	}

	return nil
}

// installDependencies installs dependencies for the server
func (m *Manager) installDependencies(serverDir string) error {
	// Check for package.json (Node.js)
	if _, err := os.Stat(filepath.Join(serverDir, "package.json")); err == nil {
		return m.runCommand(serverDir, "npm", "install")
	}

	// Check for requirements.txt (Python)
	if _, err := os.Stat(filepath.Join(serverDir, "requirements.txt")); err == nil {
		return m.runCommand(serverDir, "pip", "install", "-r", "requirements.txt")
	}

	// Check for go.mod (Go)
	if _, err := os.Stat(filepath.Join(serverDir, "go.mod")); err == nil {
		return m.runCommand(serverDir, "go", "mod", "download")
	}

	return nil
}

// buildServer builds the server based on the project type
func (m *Manager) buildServer(serverDir string) error {
	// Check for package.json (Node.js) - typically no build needed for MCP servers
	if _, err := os.Stat(filepath.Join(serverDir, "package.json")); err == nil {
		// Check if there's a build script
		if m.hasNpmScript(serverDir, "build") {
			return m.runCommand(serverDir, "npm", "run", "build")
		}
		return nil
	}

	// Check for requirements.txt (Python) - no build needed
	if _, err := os.Stat(filepath.Join(serverDir, "requirements.txt")); err == nil {
		return nil
	}

	// Check for go.mod (Go) - build the binary
	if _, err := os.Stat(filepath.Join(serverDir, "go.mod")); err == nil {
		return m.runCommand(serverDir, "go", "build", "-o", "server", ".")
	}

	// Check for Cargo.toml (Rust)
	if _, err := os.Stat(filepath.Join(serverDir, "Cargo.toml")); err == nil {
		return m.runCommand(serverDir, "cargo", "build", "--release")
	}

	return nil
}

// hasNpmScript checks if a package.json has a specific script
func (m *Manager) hasNpmScript(serverDir, script string) bool {
	packageJsonPath := filepath.Join(serverDir, "package.json")
	data, err := os.ReadFile(packageJsonPath)
	if err != nil {
		return false
	}

	var packageJson map[string]interface{}
	if err := json.Unmarshal(data, &packageJson); err != nil {
		return false
	}

	scripts, ok := packageJson["scripts"].(map[string]interface{})
	if !ok {
		return false
	}

	_, exists := scripts[script]
	return exists
}

// determineExecution determines the command and arguments to run the server
func (m *Manager) determineExecution(serverDir string) (string, []string, map[string]string, error) {
	env := make(map[string]string)

	// Check for package.json (Node.js)
	if _, err := os.Stat(filepath.Join(serverDir, "package.json")); err == nil {
		// Look for main entry point
		packageJsonPath := filepath.Join(serverDir, "package.json")
		data, err := os.ReadFile(packageJsonPath)
		if err != nil {
			return "", nil, nil, err
		}

		var packageJson map[string]interface{}
		if err := json.Unmarshal(data, &packageJson); err != nil {
			return "", nil, nil, err
		}

		// Check for bin field first
		if bin, ok := packageJson["bin"].(map[string]interface{}); ok {
			for _, binPath := range bin {
				if binPathStr, ok := binPath.(string); ok {
					return "node", []string{filepath.Join(serverDir, binPathStr)}, env, nil
				}
			}
		}

		// Check for main field
		if main, ok := packageJson["main"].(string); ok {
			return "node", []string{filepath.Join(serverDir, main)}, env, nil
		}

		// Default to index.js
		return "node", []string{filepath.Join(serverDir, "index.js")}, env, nil
	}

	// Check for Python
	if _, err := os.Stat(filepath.Join(serverDir, "requirements.txt")); err == nil {
		// Look for main.py or __main__.py
		if _, err := os.Stat(filepath.Join(serverDir, "main.py")); err == nil {
			return "python", []string{filepath.Join(serverDir, "main.py")}, env, nil
		}
		if _, err := os.Stat(filepath.Join(serverDir, "__main__.py")); err == nil {
			return "python", []string{filepath.Join(serverDir, "__main__.py")}, env, nil
		}
		// Look for setup.py and try to find entry points
		return "python", []string{"-m", filepath.Base(serverDir)}, env, nil
	}

	// Check for Go binary
	if _, err := os.Stat(filepath.Join(serverDir, "go.mod")); err == nil {
		binaryPath := filepath.Join(serverDir, "server")
		if _, err := os.Stat(binaryPath); err == nil {
			return binaryPath, []string{}, env, nil
		}
	}

	// Check for Rust binary
	if _, err := os.Stat(filepath.Join(serverDir, "Cargo.toml")); err == nil {
		// Read Cargo.toml to get binary name
		cargoTomlPath := filepath.Join(serverDir, "Cargo.toml")
		data, err := os.ReadFile(cargoTomlPath)
		if err == nil {
			// Simple parsing - look for [package] name
			lines := strings.Split(string(data), "\n")
			for _, line := range lines {
				if strings.HasPrefix(strings.TrimSpace(line), "name = ") {
					name := strings.Trim(strings.TrimPrefix(strings.TrimSpace(line), "name = "), "\"")
					binaryPath := filepath.Join(serverDir, "target", "release", name)
					if _, err := os.Stat(binaryPath); err == nil {
						return binaryPath, []string{}, env, nil
					}
				}
			}
		}
	}

	return "", nil, nil, fmt.Errorf("could not determine execution method for server")
}

// runCommand executes a command in a specific directory
func (m *Manager) runCommand(dir, command string, args ...string) error {
	cmd := exec.Command(command, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// RemoveServer removes a specific MCP server version
func (m *Manager) RemoveServer(name, version string) error {
	serverDir := filepath.Join(m.dataDir, name, version)

	if _, err := os.Stat(serverDir); os.IsNotExist(err) {
		return fmt.Errorf("server %s@%s is not installed", name, version)
	}

	if err := os.RemoveAll(serverDir); err != nil {
		return fmt.Errorf("failed to remove server: %w", err)
	}

	// Remove parent directory if empty
	parentDir := filepath.Join(m.dataDir, name)
	if isEmpty, _ := isDirEmpty(parentDir); isEmpty {
		os.Remove(parentDir)
	}

	// Remove server from agent configurations
	if err := m.RemoveServerFromAgentConfigs(name); err != nil {
		fmt.Printf("Warning: Failed to remove server from agent configurations: %v\n", err)
	}

	return nil
}

// ListInstalledServers returns a list of installed servers
func (m *Manager) ListInstalledServers() ([]MCPServer, error) {
	var servers []MCPServer

	entries, err := os.ReadDir(m.dataDir)
	if err != nil {
		if os.IsNotExist(err) {
			return servers, nil
		}
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		serverName := entry.Name()
		serverDir := filepath.Join(m.dataDir, serverName)

		versions, err := os.ReadDir(serverDir)
		if err != nil {
			continue
		}

		for _, versionEntry := range versions {
			if !versionEntry.IsDir() {
				continue
			}

			servers = append(servers, MCPServer{
				Name:        serverName,
				Version:     versionEntry.Name(),
				InstallPath: filepath.Join(serverDir, versionEntry.Name()),
				Installed:   true,
			})
		}
	}

	return servers, nil
}

// UpdateServer updates a server to the latest version
func (m *Manager) UpdateServer(name, repository string) error {
	// Get latest version from repository
	latestVersion, err := m.getLatestVersion(repository)
	if err != nil {
		return fmt.Errorf("failed to get latest version: %w", err)
	}

	// Install the latest version
	_, err = m.InstallServer(name, latestVersion, repository)
	return err
}

// getLatestVersion gets the latest version from a git repository
func (m *Manager) getLatestVersion(repoURL string) (string, error) {
	// This is a simplified implementation
	// In a real implementation, you'd query the git repository for tags
	return "latest", nil
	return "latest", nil
}

// InstallFromConfig installs all servers specified in the project config
func (m *Manager) InstallFromConfig(configPath string) error {
	config, err := m.LoadProjectConfig(configPath)
	if err != nil {
		return err
	}

	for _, server := range config.Servers {
		if server.Repository == "" {
			return fmt.Errorf("repository not specified for server %s", server.Name)
		}

		version := server.Version
		if version == "" {
			version = "latest"
		}

		fmt.Printf("Installing %s@%s...\n", server.Name, version)
		installedServer, err := m.InstallServer(server.Name, version, server.Repository)
		if err != nil {
			if strings.Contains(err.Error(), "already installed") {
				fmt.Printf("Server %s@%s is already installed\n", server.Name, version)
				// Still need to patch agent configs for already installed servers
				installedServer = &MCPServer{
					Name:       server.Name,
					Version:    version,
					Repository: server.Repository,
					Command:    server.Command,
					Args:       server.Args,
					Env:        server.Env,
				}

				// Try to determine execution details if not provided
				if installedServer.Command == "" {
					serverDir := filepath.Join(m.dataDir, server.Name, version)
					command, args, env, execErr := m.determineExecution(serverDir)
					if execErr == nil {
						installedServer.Command = command
						installedServer.Args = args
						installedServer.Env = env
					}
				}
			} else {
				return err
			}
		} else {
			fmt.Printf("Successfully installed %s@%s\n", server.Name, version)
		}

		// Patch agent configurations for all detected agents
		if err := m.PatchAgentConfigs(installedServer); err != nil {
			fmt.Printf("Warning: Failed to configure server %s for agents: %v\n", server.Name, err)
		}
	}

	return nil
}

// InstallFromConfigForAgent installs all servers specified in the project config for a specific agent
func (m *Manager) InstallFromConfigForAgent(configPath string, agentType AgentType) error {
	config, err := m.LoadProjectConfig(configPath)
	if err != nil {
		return err
	}

	for _, server := range config.Servers {
		if server.Repository == "" {
			return fmt.Errorf("repository not specified for server %s", server.Name)
		}

		version := server.Version
		if version == "" {
			version = "latest"
		}

		fmt.Printf("Installing %s@%s for %s agent...\n", server.Name, version, agentType)
		_, err := m.InstallServer(server.Name, version, server.Repository)
		if err != nil {
			if strings.Contains(err.Error(), "already installed") {
				fmt.Printf("Server %s@%s is already installed\n", server.Name, version)
				// Still need to configure for the specific agent
			} else {
				return err
			}
		} else {
			fmt.Printf("Successfully installed %s@%s\n", server.Name, version)
		}

		// Configure for specific agent only
		installedServer := &MCPServer{
			Name:       server.Name,
			Version:    version,
			Repository: server.Repository,
			Command:    server.Command,
			Args:       server.Args,
			Env:        server.Env,
		}

		// If we don't have execution details, try to determine them
		if installedServer.Command == "" {
			serverDir := filepath.Join(m.dataDir, server.Name, version)
			command, args, env, err := m.determineExecution(serverDir)
			if err == nil {
				installedServer.Command = command
				installedServer.Args = args
				installedServer.Env = env
			}
		}

		if err := m.agentConfigManager.AddServerToAgent(agentType, installedServer); err != nil {
			return fmt.Errorf("failed to configure server %s for %s: %w", server.Name, agentType, err)
		}
		fmt.Printf("✓ Configured %s for %s agent\n", server.Name, agentType)
	}

	return nil
}

// InstallServerAndAddToConfig installs a server and adds it to the mcpv.json configuration
func (m *Manager) InstallServerAndAddToConfig(name, version, repository, configPath string) error {
	server, err := m.InstallServer(name, version, repository)
	if err != nil {
		return err
	}

	// Load existing config
	config, err := m.LoadProjectConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Check if server already exists in config
	for i, existingServer := range config.Servers {
		if existingServer.Name == name && existingServer.Version == version {
			// Update existing server with execution details
			config.Servers[i] = *server
			if err := m.SaveProjectConfig(config, configPath); err != nil {
				return err
			}
			// Patch agent configurations
			return m.PatchAgentConfigs(server)
		}
	}

	// Add new server to config
	config.Servers = append(config.Servers, *server)
	if err := m.SaveProjectConfig(config, configPath); err != nil {
		return err
	}

	// Patch agent configurations
	return m.PatchAgentConfigs(server)
}

// InstallServerAndAddToConfigForAgent installs a server and adds it to the mcpv.json configuration for a specific agent
func (m *Manager) InstallServerAndAddToConfigForAgent(name, version, repository, configPath string, agentType AgentType) error {
	// Install the server
	server, err := m.InstallServer(name, version, repository)
	if err != nil {
		return err
	}

	// Load existing config
	config, err := m.LoadProjectConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Check if server already exists in config
	for i, existingServer := range config.Servers {
		if existingServer.Name == name && existingServer.Version == version {
			// Update existing server with execution details
			config.Servers[i] = *server
			if err := m.SaveProjectConfig(config, configPath); err != nil {
				return err
			}
			// Configure for specific agent only
			return m.agentConfigManager.AddServerToAgent(agentType, server)
		}
	}

	// Add new server to config
	config.Servers = append(config.Servers, *server)
	if err := m.SaveProjectConfig(config, configPath); err != nil {
		return err
	}

	// Configure for specific agent only
	return m.agentConfigManager.AddServerToAgent(agentType, server)
}

// isDirEmpty checks if a directory is empty
func isDirEmpty(dir string) (bool, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false, err
	}
	return len(entries) == 0, nil
}

// ParseServerSpec parses a server specification like "server@1.0.0"
func ParseServerSpec(spec string) (name, version string) {
	parts := strings.Split(spec, "@")
	name = parts[0]
	if len(parts) > 1 {
		version = parts[1]
	}
	return
}

// ValidateVersion validates a semantic version
func ValidateVersion(version string) error {
	if version == "" || version == "latest" {
		return nil
	}
	_, err := semver.NewVersion(version)
	return err
}

// PatchAgentConfigs patches agent configurations with the installed MCP server
func (m *Manager) PatchAgentConfigs(server *MCPServer) error {
	if m.agentConfigManager == nil {
		return fmt.Errorf("agent config manager not initialized")
	}

	// Detect available agents
	availableAgents := m.agentConfigManager.DetectAvailableAgents()
	if len(availableAgents) == 0 {
		fmt.Printf("No supported AI agents detected. Server %s installed but not configured for any agents.\n", server.Name)
		return nil
	}

	fmt.Printf("Configuring server %s for detected agents...\n", server.Name)

	// Add server to all available agents
	var errors []string
	for _, agentType := range availableAgents {
		if err := m.agentConfigManager.AddServerToAgent(agentType, server); err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", agentType, err))
		} else {
			fmt.Printf("✓ Added %s to %s configuration\n", server.Name, agentType)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to configure server for some agents: %v", errors)
	}

	return nil
}

// RemoveServerFromAgentConfigs removes an MCP server from agent configurations
func (m *Manager) RemoveServerFromAgentConfigs(serverName string) error {
	if m.agentConfigManager == nil {
		return fmt.Errorf("agent config manager not initialized")
	}

	// Detect available agents
	availableAgents := m.agentConfigManager.DetectAvailableAgents()
	if len(availableAgents) == 0 {
		return nil // No agents to remove from
	}

	fmt.Printf("Removing server %s from agent configurations...\n", serverName)

	// Remove server from all available agents
	var errors []string
	for _, agentType := range availableAgents {
		if err := m.agentConfigManager.RemoveServerFromAgent(agentType, serverName); err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", agentType, err))
		} else {
			fmt.Printf("✓ Removed %s from %s configuration\n", serverName, agentType)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to remove server from some agents: %v", errors)
	}

	return nil
}

// ListAgentConfigurations lists the current agent configurations
func (m *Manager) ListAgentConfigurations() error {
	if m.agentConfigManager == nil {
		return fmt.Errorf("agent config manager not initialized")
	}

	availableAgents := m.agentConfigManager.DetectAvailableAgents()
	if len(availableAgents) == 0 {
		fmt.Println("No supported AI agents detected.")
		return nil
	}

	fmt.Println("Detected AI agents:")
	for _, agentType := range availableAgents {
		config, exists := m.agentConfigManager.configs[string(agentType)]
		if !exists {
			continue
		}

		configPath, err := config.GetConfigPath()
		if err != nil {
			fmt.Printf("- %s: Error getting config path: %v\n", agentType, err)
			continue
		}

		fmt.Printf("- %s: %s\n", agentType, configPath)
	}

	return nil
}

// AddServerToAgent adds an MCP server to a specific agent's configuration
func (m *Manager) AddServerToAgent(agentType AgentType, server *MCPServer) error {
	if m.agentConfigManager == nil {
		return fmt.Errorf("agent config manager not initialized")
	}

	return m.agentConfigManager.AddServerToAgent(agentType, server)
}

// AddServerToAgentWithLocal adds an MCP server to a specific agent's configuration with local preference
func (m *Manager) AddServerToAgentWithLocal(agentType AgentType, server *MCPServer, useLocal bool) error {
	if m.agentConfigManager == nil {
		return fmt.Errorf("agent config manager not initialized")
	}

	return m.agentConfigManager.AddServerToAgentWithLocal(agentType, server, useLocal)
}

// InstallServerAndAddToConfigForAgentWithLocal installs a server and adds it to the mcpv.json configuration for a specific agent with local preference
func (m *Manager) InstallServerAndAddToConfigForAgentWithLocal(name, version, repository, configPath string, agentType AgentType, useLocal bool) error {
	// Install the server
	server, err := m.InstallServer(name, version, repository)
	if err != nil {
		return err
	}

	// Load existing config
	config, err := m.LoadProjectConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Check if server already exists in config
	for i, existingServer := range config.Servers {
		if existingServer.Name == name && existingServer.Version == version {
			// Update existing server with execution details
			config.Servers[i] = *server
			if err := m.SaveProjectConfig(config, configPath); err != nil {
				return err
			}
			// Configure for specific agent with local preference
			return m.AddServerToAgentWithLocal(agentType, server, useLocal)
		}
	}

	// Add new server to config
	config.Servers = append(config.Servers, *server)
	if err := m.SaveProjectConfig(config, configPath); err != nil {
		return err
	}

	// Configure for specific agent with local preference
	return m.AddServerToAgentWithLocal(agentType, server, useLocal)
}

// InstallFromConfigForAgentWithLocal installs all servers specified in the project config for a specific agent with local preference
func (m *Manager) InstallFromConfigForAgentWithLocal(configPath string, agentType AgentType, useLocal bool) error {
	config, err := m.LoadProjectConfig(configPath)
	if err != nil {
		return err
	}

	for _, server := range config.Servers {
		if server.Repository == "" {
			return fmt.Errorf("repository not specified for server %s", server.Name)
		}

		version := server.Version
		if version == "" {
			version = "latest"
		}

		configType := "local"
		if !useLocal {
			configType = "global"
		}
		fmt.Printf("Installing %s@%s for %s agent (%s config)...\n", server.Name, version, agentType, configType)
		_, err := m.InstallServer(server.Name, version, server.Repository)
		if err != nil {
			if strings.Contains(err.Error(), "already installed") {
				fmt.Printf("Server %s@%s is already installed\n", server.Name, version)
				// Still need to configure for the specific agent
			} else {
				return err
			}
		} else {
			fmt.Printf("Successfully installed %s@%s\n", server.Name, version)
		}

		// Configure for specific agent with local preference
		installedServer := &MCPServer{
			Name:       server.Name,
			Version:    version,
			Repository: server.Repository,
			Command:    server.Command,
			Args:       server.Args,
			Env:        server.Env,
		}

		// If we don't have execution details, try to determine them
		if installedServer.Command == "" {
			serverDir := filepath.Join(m.dataDir, server.Name, version)
			command, args, env, err := m.determineExecution(serverDir)
			if err == nil {
				installedServer.Command = command
				installedServer.Args = args
				installedServer.Env = env
			}
		}

		if err := m.AddServerToAgentWithLocal(agentType, installedServer, useLocal); err != nil {
			return fmt.Errorf("failed to configure server %s for %s: %w", server.Name, agentType, err)
		}
		fmt.Printf("✓ Configured %s for %s agent (%s config)\n", server.Name, agentType, configType)
	}

	return nil
}

// GetAgentConfigManager returns the agent configuration manager
func (m *Manager) GetAgentConfigManager() *DynamicAgentConfigManager {
	return m.agentConfigManager
}
