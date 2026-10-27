package plugins

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// settingsNodeFor returns the YAML node of the given row type settings
func settingsNodeFor(t *testing.T, settingsYAML string) *yaml.Node {
	var node yaml.Node
	require.NoError(t, yaml.Unmarshal([]byte(settingsYAML), &node))
	// unwrap the document node
	return node.Content[0]
}
