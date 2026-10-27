package dynamotd

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"
)

// RowType generates rows of the MOTD - usually one, but it may be none (f.e. when there is nothing
// to warn about) or more than one (f.e. one per filesystem).
type RowType struct {
	// Label is used for the row shown when generating fails or doesn't finish in time.
	Label string
	// Generate should return as soon as possible after ctx is done, even if with partial info.
	Generate func(ctx context.Context) []Row
}

var registry = map[string]RowType{}

// Register makes a row type available under the given name(s) in the config file.
// It is meant to be called from the plugins' init() functions.
func Register(rowType RowType, names ...string) {
	for _, name := range names {
		if _, exists := registry[name]; exists {
			panic(fmt.Sprintf("row type '%s' registered twice", name))
		}
		registry[name] = rowType
	}
}

// Lookup returns the row types with the given names.
func Lookup(names []string) ([]RowType, error) {
	rowTypes := make([]RowType, 0, len(names))
	for _, name := range names {
		rowType, ok := registry[name]
		if !ok {
			return nil, fmt.Errorf("unknown row type '%s' - available ones are: %s", name, strings.Join(Names(), ", "))
		}
		rowTypes = append(rowTypes, rowType)
	}
	return rowTypes, nil
}

// Names returns the names of all the registered row types, except the empty line's empty name.
func Names() []string {
	var names []string
	for name := range registry {
		if name != "" {
			names = append(names, name)
		}
	}
	slices.Sort(names)
	return names
}

// how long to wait after the timeout for the row types that return partial info when their context is done
const gracePeriod = 100 * time.Millisecond

// Generate runs all the row types concurrently, so that the slowest one (f.e. getting the external IP)
// determines the total time, not the sum of all of them. Row types that don't finish within the timeout
// are shown as timed out, so that a server with f.e. broken DNS doesn't block the login.
// Row types that panic are shown as errors.
func Generate(rowTypes []RowType, timeout time.Duration) []Row {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	type result struct {
		index int
		rows  []Row
	}
	results := make(chan result, len(rowTypes))

	for i, rowType := range rowTypes {
		go func() {
			defer func() {
				if r := recover(); r != nil {
					results <- result{i, []Row{ErrorRow(rowType.Label, fmt.Errorf("%v", r))}}
				}
			}()
			results <- result{i, rowType.Generate(ctx)}
		}()
	}

	rowsByType := make([][]Row, len(rowTypes))
	finished := make([]bool, len(rowTypes))
	deadline := time.NewTimer(timeout + gracePeriod)
	defer deadline.Stop()

	for range rowTypes {
		select {
		case r := <-results:
			rowsByType[r.index] = r.rows
			finished[r.index] = true
		case <-deadline.C:
			for i, rowType := range rowTypes {
				if !finished[i] {
					rowsByType[i] = []Row{TimedOutRow(rowType.Label)}
				}
			}
			return flatten(rowsByType)
		}
	}
	return flatten(rowsByType)
}

func flatten(rowsByType [][]Row) []Row {
	var rows []Row
	for _, typeRows := range rowsByType {
		rows = append(rows, typeRows...)
	}
	return rows
}

// SingleRow adapts a function generating exactly one row, not using the context, to RowType.Generate.
func SingleRow(generate func() Row) func(ctx context.Context) []Row {
	return func(context.Context) []Row {
		return []Row{generate()}
	}
}
