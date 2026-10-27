package dynamotd

import (
	"fmt"
	"sync"
)

// RowFunc generates a single row of the MOTD.
type RowFunc func() Row

var registry = map[string]RowFunc{}

// Register makes a row type available under the given name(s) in the config file.
// It is meant to be called from the plugins' init() functions.
func Register(rowFunc RowFunc, names ...string) {
	for _, name := range names {
		if _, exists := registry[name]; exists {
			panic(fmt.Sprintf("row type '%s' registered twice", name))
		}
		registry[name] = rowFunc
	}
}

// Lookup returns the row generating functions for the given row type names.
func Lookup(names []string) ([]RowFunc, error) {
	rowFuncs := make([]RowFunc, 0, len(names))
	for _, name := range names {
		rowFunc, ok := registry[name]
		if !ok {
			return nil, fmt.Errorf("unknown row type '%s' - check for typos", name)
		}
		rowFuncs = append(rowFuncs, rowFunc)
	}
	return rowFuncs, nil
}

// Generate runs all the row generating functions concurrently, so that the slowest one
// (f.e. getting the external IP) determines the total time, not the sum of all of them.
// A row whose function panics is replaced with an error row.
func Generate(rowFuncs []RowFunc) []Row {
	rows := make([]Row, len(rowFuncs))

	var wg sync.WaitGroup
	for i, rowFunc := range rowFuncs {
		wg.Go(func() {
			defer func() {
				if r := recover(); r != nil {
					rows[i] = ErrorRow("Error", fmt.Errorf("%v", r))
				}
			}()
			rows[i] = rowFunc()
		})
	}
	wg.Wait()

	return rows
}
