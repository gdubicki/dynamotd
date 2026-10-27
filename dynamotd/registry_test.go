package dynamotd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLookupUnknown(t *testing.T) {
	_, err := Lookup([]string{"no-such-row"})
	assert.ErrorContains(t, err, "unknown row type 'no-such-row'")
}

func TestGenerateKeepsOrderAndRecoversPanics(t *testing.T) {
	rows := Generate([]RowFunc{
		func() Row { return Row{Label: SingleColorLabel("first")} },
		func() Row { panic("boom") },
		func() Row { return Row{Label: SingleColorLabel("third")} },
	})

	require.Len(t, rows, 3)
	assert.Equal(t, "first", rows[0].Label.String())
	assert.Equal(t, "can't get: boom", rows[1].Value.String())
	assert.Equal(t, "third", rows[2].Label.String())
}
