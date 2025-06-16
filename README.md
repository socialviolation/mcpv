# mcpv - MCP Server Version Manager

`mcpv` is a CLI tool for managing Model Context Protocol (MCP) servers on your local machine. It provides functionality to install, update, and delete MCP servers with support for multiple versions of the same server.

## Features

- **Version Management**: Install and manage multiple versions of the same MCP server
- **Project Configuration**: Use `mcpv.json` files to specify server dependencies per project
- **XDG Compliance**: Installs servers to `XDG_DATA_HOME` (or `~/.local/share/mcpv` by default)
- **Multiple Languages**: Supports Node.js, Python, and Go MCP servers
- **Git Integration**: Clone servers directly from Git repositories

## Installation

Build from source:

```bash
git clone https://github.com/socialviolation/mcpv.git
cd mcpv
go build -o mcpv
```

## Usage

### Initialize a Project

Create a new `mcpv.json` configuration file in your project directory:

```bash
mcpv init
```

This creates a sample configuration file:

```json
{
  "servers": [
    {
      "name": "example-server",
      "version": "1.0.0",
      "repository": "https://github.com/example/mcp-server.git"
    },
    {
      "name": "another-server",
      "version": "latest",
      "repository": "https://github.com/example/another-mcp-server.git"
    }
  ]
}
```

### Install Servers

Install all servers specified in `mcpv.json`:

```bash
mcpv install
```

Install from a specific config file:

```bash
mcpv install --config path/to/mcpv.json
```

### List Servers

List all installed servers:

```bash
mcpv list
```

List servers configured in the current project:

```bash
mcpv list --project
```

### Update Servers

Update all servers from `mcpv.json` to their latest versions:

```bash
mcpv update
```

### Remove Servers

Remove a specific server version:

```bash
mcpv remove server@1.0.0
```

Remove all versions of a server:

```bash
mcpv remove server
```

## Configuration

### mcpv.json Schema

```json
{
  "servers": [
    {
      "name": "server-name",           // Required: Name of the server
      "version": "1.0.0",              // Optional: Version (defaults to "latest")
      "repository": "https://..."      // Required: Git repository URL
    }
  ]
}
```

### Storage Location

Servers are installed to:
- `$XDG_DATA_HOME/mcpv/` if `XDG_DATA_HOME` is set
- `~/.local/share/mcpv/` otherwise

Directory structure:
```
~/.local/share/mcpv/
├── server-name/
│   ├── 1.0.0/
│   ├── 1.1.0/
│   └── latest/
└── another-server/
    └── 2.0.0/
```

## Supported Server Types

### Node.js Servers
- Detects `package.json`
- Runs `npm install` automatically

### Python Servers
- Detects `requirements.txt`
- Runs `pip install -r requirements.txt` automatically

### Go Servers
- Detects `go.mod`
- Runs `go mod download` automatically

## Examples

### Basic Workflow

1. Initialize a new project:
   ```bash
   mcpv init
   ```

2. Edit `mcpv.json` to specify your required servers:
   ```json
   {
     "servers": [
       {
         "name": "filesystem-server",
         "version": "1.2.0",
         "repository": "https://github.com/modelcontextprotocol/servers.git"
       }
     ]
   }
   ```

3. Install the servers:
   ```bash
   mcpv install
   ```

4. List installed servers:
   ```bash
   mcpv list
   ```

### Managing Multiple Versions

```bash
# Install different versions of the same server
mcpv install server@1.0.0
mcpv install server@2.0.0

# List all versions
mcpv list

# Remove specific version
mcpv remove server@1.0.0
```

## Development

### Building

```bash
go build -o mcpv
```

### Testing

```bash
go test ./...
```

### Dependencies

- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [go-git](https://github.com/go-git/go-git) - Git operations
- [semver](https://github.com/Masterminds/semver) - Semantic versioning

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## Roadmap

- [ ] Server registry support
- [ ] Automatic dependency resolution
- [ ] Configuration validation
- [ ] Shell completion
- [ ] Docker support
- [ ] Server health checks
- [ ] Backup/restore functionality