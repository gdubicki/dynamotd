package dynamotd

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLookupUnknown(t *testing.T) {
	_, err := Lookup([]string{"no-such-row"})
	assert.ErrorContains(t, err, "unknown row type 'no-such-row'")
}

func rowType(label string, generate func(ctx context.Context) []Row) RowType {
	return RowType{Label: label, Generate: generate}
}

func labeled(label string) []Row {
	return []Row{{Label: SingleColorLabel(label)}}
}

func TestGenerateKeepsOrderAndRecoversPanics(t *testing.T) {
	rows := Generate([]RowType{
		rowType("first", func(context.Context) []Row { return labeled("first") }),
		rowType("second", func(context.Context) []Row { panic("boom") }),
		rowType("third", func(context.Context) []Row { return labeled("third") }),
	}, time.Second)

	require.Len(t, rows, 3)
	assert.Equal(t, "first", rows[0].Label.String())
	assert.Equal(t, "second", rows[1].Label.String())
	assert.Equal(t, "can't get: boom", rows[1].Value.String())
	assert.Equal(t, "third", rows[2].Label.String())
}

func TestGenerateNoneOrManyRows(t *testing.T) {
	rows := Generate([]RowType{
		rowType("none", func(context.Context) []Row { return nil }),
		rowType("many", func(context.Context) []Row { return append(labeled("a"), labeled("b")...) }),
	}, time.Second)

	require.Len(t, rows, 2)
	assert.Equal(t, "a", rows[0].Label.String())
	assert.Equal(t, "b", rows[1].Label.String())
}

func TestGenerateTimeout(t *testing.T) {
	block := make(chan struct{})
	defer close(block)

	start := time.Now()
	rows := Generate([]RowType{
		rowType("fast", func(context.Context) []Row { return labeled("fast") }),
		// ignores the context, like a DNS lookup without a timeout
		rowType("hanging", func(context.Context) []Row { <-block; return labeled("never") }),
		// returns partial info when the context is done
		rowType("partial", func(ctx context.Context) []Row { <-ctx.Done(); return labeled("partial info") }),
	}, 50*time.Millisecond)

	assert.Less(t, time.Since(start), 50*time.Millisecond+gracePeriod+100*time.Millisecond)
	require.Len(t, rows, 3)
	assert.Equal(t, "fast", rows[0].Label.String())
	assert.Equal(t, "hanging", rows[1].Label.String())
	assert.Equal(t, "timed out", rows[1].Value.String())
	assert.Equal(t, "partial info", rows[2].Label.String())
}
