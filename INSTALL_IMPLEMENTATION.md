# Install Logic Implementation

## Overview

I have successfully implemented the install logic for the `mcpv install` command that supports installing specific MCP servers with the following features:

## Key Features Implemented

### 1. Server Installation with Repository URL
- Servers are cloned to `$XDG_DATA_HOME/mcpv/{server_name}/{version}` (or `~/.local/share/mcpv/` on macOS/Linux)
- Supports installing specific server versions using `mcpv install server@version --repo <repository-url>`
- Automatically detects project type and builds the server appropriately

### 2. Multi-Language Support
The implementation supports multiple programming languages:

#### Node.js Projects
- Runs `npm install` for dependencies
- Runs `npm run build` if a build script exists
- Determines execution via `package.json` bin field, main field, or defaults to `index.js`

#### Python Projects  
- Runs `pip install -r requirements.txt` for dependencies
- No build step required
- Determines execution via `main.py`, `__main__.py`, or module execution

#### Go Projects
- Runs `go mod download` for dependencies  
- Builds binary with `go build -o server .`
- Executes the built binary directly

#### Rust Projects
- Builds with `cargo build --release`
- Determines binary name from `Cargo.toml`
- Executes the release binary

### 3. Configuration Management
- Automatically adds installed servers to `mcpv.json` with execution details
- Includes command, arguments, and environment variables for each server
- Updates existing server entries if they already exist

### 4. Enhanced MCPServer Structure
Extended the `MCPServer` struct to include execution configuration:
```go
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
```

## Usage Examples

### Install from mcpv.json
```bash
mcpv install
```

### Install specific server with repository URL
```bash
mcpv install my-server@1.0.0 --repo https://github.com/user/my-mcp-server.git
```

### Install latest version
```bash
mcpv install my-server --repo https://github.com/user/my-mcp-server.git
```

## Implementation Details

### New Methods Added

1. **`InstallServer(name, version, repository string) (*MCPServer, error)`**
   - Clones repository to appropriate directory
   - Installs dependencies based on project type
   - Builds the server if required
   - Determines execution configuration
   - Returns complete server configuration

2. **`buildServer(serverDir string) error`**
   - Detects project type and runs appropriate build commands
   - Handles Node.js, Python, Go, and Rust projects

3. **`determineExecution(serverDir string) (string, []string, map[string]string, error)`**
   - Analyzes project structure to determine how to execute the server
   - Returns command, arguments, and environment variables

4. **`InstallServerAndAddToConfig(name, version, repository, configPath string) error`**
   - Installs server and automatically adds it to mcpv.json
   - Updates existing entries or adds new ones

### Directory Structure
Servers are installed in the following structure:
```
$XDG_DATA_HOME/mcpv/
├── server-name/
│   ├── 1.0.0/
│   │   ├── (cloned repository contents)
│   │   └── (built artifacts)
│   └── 1.1.0/
│       ├── (cloned repository contents)
│       └── (built artifacts)
└── another-server/
    └── latest/
        ├── (cloned repository contents)
        └── (built artifacts)
```

## Error Handling
- Validates repository URLs are provided for specific server installation
- Handles authentication errors gracefully
- Cleans up partially installed servers on failure
- Provides clear error messages and usage instructions

## Integration with AI Tools
The generated mcpv.json configuration with execution details can be used to generate MCP configurations for various AI integrations including:
- Claude Desktop
- Cline
- Windsurf  
- ChatGPT
- And other MCP-compatible tools

This implementation provides a solid foundation for managing MCP servers with proper building, installation, and configuration management.