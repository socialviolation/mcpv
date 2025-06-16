package manager

import (
	_ "embed"
)

//go:embed agents.yaml
var embeddedAgentsYAML string

// getEmbeddedAgentsYAML returns the embedded agents YAML content
func getEmbeddedAgentsYAML() string {
	return embeddedAgentsYAML
}
