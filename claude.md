# mcpv - MCP Version Manager

## Project Overview

**mcpv** is a command-line tool built in Go using the Cobra CLI framework for managing Model Context Protocol (MCP) servers. It provides installation, version management, and configuration capabilities for MCP servers from various sources.

## Core Purpose

- Install MCP servers locally without manual cloning and building
- Manage updates and versions of installed MCP servers
- Provide a centralized way to configure and manage MCP server installations
- Support multiple installation sources (GitHub, npm, PyPI)

## Architecture

### Technology Stack
- **Language**: Go
- **CLI Framework**: Cobra CLI
- **Configuration**: Viper (YAML-based)
- **Package Management**: Go modules

### Project Structure
```
mcpv/
├── cmd/                    # Cobra command definitions
│   ├── root.go            # Root command and global config
│   ├── install.go         # Install command
│   ├── list.go            # List installed servers
│   ├── update.go          # Update servers  
│   ├── remove.go          # Remove/uninstall servers
│   └── config.go          # Configuration management
├── internal/              # Internal packages
│   ├── types/             # Data structures and types
│   ├── installer/         # Installation logic
│   ├── manager/           # Server management
│   ├── updater/           # Update functionality
│   └── config/            # Configuration handling
├── main.go               # Entry point
├── go.mod                # Go module definition
└── README.md             # Documentation
```

## Key Components

### Commands
- `mcpv install <server>` - Install MCP servers from various sources
- `mcpv list` - Show installed servers with versions and status
- `mcpv update [server|--all]` - Update specific or all servers
- `mcpv remove <server>` - Uninstall servers
- `mcpv config [init|show]` - Manage configuration

### Installation Sources
1. **GitHub Repositories** - Clone and build from Git repos
2. **npm Packages** - Install Node.js-based MCP servers
3. **PyPI Packages** - Install Python-based MCP servers
4. **Known Server Registry** - Pre-configured popular servers

### Configuration
- Config file: `~/.mcpv.yaml`
- Install directory: `~/.mcpv/servers/`
- Server-specific configurations with environment variables and arguments

## Data Structures

### MCPServer
```go
type MCPServer struct {
    Name        string    `json:"name"`
    Version     string    `json:"version"`
    Source      string    `json:"source"`
    InstallPath string    `json:"install_path"`
    Status      string    `json:"status"`
    InstalledAt time.Time `json:"installed_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}
```

### Config
```go
type Config struct {
    InstallDir   string                    `yaml:"install_dir"`
    Registries   map[string]string         `yaml:"registries"`
    Servers      map[string]ServerConfig   `yaml:"servers"`
    DefaultArgs  map[string][]string       `yaml:"default_args"`
}
```

## Implementation Guidelines

### Code Style
- Follow Go conventions and idioms
- Use Cobra's command structure and patterns
- Implement proper error handling with wrapped errors
- Use structured logging where appropriate
- Keep commands focused and single-purpose

### Error Handling
- Return meaningful error messages to users
- Use `fmt.Errorf` with error wrapping
- Validate inputs before processing
- Handle network failures gracefully

### Testing Considerations
- Unit tests for installer logic
- Integration tests for command execution
- Mock external dependencies (GitHub API, npm, etc.)
- Test configuration file handling

## Development Priorities

### Phase 1 (MVP)
1. Basic command structure with Cobra
2. Configuration file handling
3. GitHub repository installation
4. List and remove functionality
5. Known servers registry

### Phase 2 (Extended Sources)
1. npm package installation
2. PyPI package installation
3. Version management and updates
4. Enhanced configuration options

### Phase 3 (Advanced Features)
1. Server health checks
2. Dependency management
3. Plugin system for custom installers
4. Interactive configuration wizard

## Known Servers Registry

Popular MCP servers to include in the built-in registry:
- `filesystem` - File system operations
- `sqlite` - SQLite database operations  
- `postgres` - PostgreSQL database operations
- `brave-search` - Brave search integration
- `git` - Git repository operations
- `github` - GitHub API integration

## Security Considerations

- Validate all URLs and package names before installation
- Sandbox installations where possible
- Verify checksums for downloaded packages
- Handle authentication for private repositories
- Sanitize user inputs for shell commands

## Dependencies

### Core Dependencies
- `github.com/spf13/cobra` - CLI framework
- `github.com/spf13/viper` - Configuration management

### Potential Additional Dependencies
- `go-git/go-git` - Git operations
- `golang.org/x/sys` - System-specific operations
- HTTP client libraries for API interactions

## Future Enhancements

- Web UI for server management
- Docker-based server isolation
- Server marketplace/discovery
- Automated testing of installed servers
- Performance monitoring and metrics
- Backup and restore functionality

## Development Notes

- Prioritize cross-platform compatibility (Windows, macOS, Linux)
- Consider using embed for built-in server registry
- Implement graceful handling of interrupted installations
- Support for air-gapped environments
- Consider integration with existing MCP client tools