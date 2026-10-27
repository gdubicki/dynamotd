package dynamotd

import (
	"errors"
	"testing"

	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
)

func TestErrorRow(t *testing.T) {
	row := ErrorRow("RAM", errors.New("boom"))
	assert.Equal(t, "RAM", row.Label.String())
	assert.Equal(t, "can't get: boom", row.Value.String())
	assert.False(t, row.isEmptyLine())
}

func TestGetStringToPrint(t *testing.T) {
	color.NoColor = false
	defer func() { color.NoColor = true }()

	// percent signs must not be interpreted as format verbs
	assert.Equal(t, "100%", getStringToPrint(ValueDescription("100%")))
	assert.Equal(t, "\x1b[31m100%\x1b[0m", getStringToPrint(ValueCritical("100%")))
}

func TestWithoutRedundantEmptyLines(t *testing.T) {
	row := func(label string) Row { return Row{Label: SingleColorLabel(label)} }
	empty := Row{}

	tests := []struct {
		name     string
		rows     []Row
		expected []Row
	}{
		{"nothing", nil, nil},
		{"no empty lines", []Row{row("a"), row("b")}, []Row{row("a"), row("b")}},
		{"single empty lines are kept", []Row{row("a"), empty, row("b")}, []Row{row("a"), empty, row("b")}},
		{"consecutive empty lines are merged", []Row{row("a"), empty, empty, empty, row("b")}, []Row{row("a"), empty, row("b")}},
		{"leading and trailing empty lines are removed", []Row{empty, row("a"), empty, empty}, []Row{row("a")}},
		{"only empty lines", []Row{empty, empty}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, withoutRedundantEmptyLines(tt.rows))
		})
	}
}
