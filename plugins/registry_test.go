package plugins

import (
	"testing"

	"github.com/gdubicki/dynamotd/dynamotd"
	"github.com/stretchr/testify/assert"
)

func TestNamesListsRegisteredRowTypes(t *testing.T) {
	names := dynamotd.Names()
	assert.Contains(t, names, "memory")
	assert.Contains(t, names, "emptyLine")
	assert.NotContains(t, names, "")
	assert.IsIncreasing(t, names)
}
