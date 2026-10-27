package dynamotd

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// RowType generates rows of the MOTD - usually one, but it may be none (f.e. when there is nothing
// to warn about) or more than one (f.e. one per filesystem).
type RowType struct {
	// Label is used for the row shown when generating fails or doesn't finish in time.
	Label string
	// Generate should return as soon as possible after ctx is done, even if with partial info.
	Generate func(ctx context.Context) []Row
}

// RowConfig is a row type to show, as configured in the config file.
type RowConfig struct {
	Name string
	// Settings of the row type, nil if not set
	Settings *yaml.Node
}

// newRowType creates a row type with the given settings
type newRowType func(settings *yaml.Node) (RowType, error)

var registry = map[string]newRowType{}

// Register makes a row type available under the given name(s) in the config file.
// It is meant to be called from the plugins' init() functions.
//
// The settings for generate are the defaults, overridden with the ones from the config file,
// so they should be a struct with yaml tags (NoSettings if there are none). If it implements
// Validator, then it's used to validate them. If it implements Labeler, then it overrides the label.
func Register[S any](label string, defaults S, generate func(ctx context.Context, settings S) []Row, names ...string) {
	newRowType := func(settingsNode *yaml.Node) (RowType, error) {
		settings, err := decodeSettings(defaults, settingsNode)
		if err != nil {
			return RowType{}, err
		}
		rowLabel := label
		if labeler, ok := any(settings).(Labeler); ok {
			rowLabel = labeler.RowLabel()
		}
		return RowType{
			Label:    rowLabel,
			Generate: func(ctx context.Context) []Row { return generate(ctx, settings) },
		}, nil
	}

	for _, name := range names {
		if _, exists := registry[name]; exists {
			panic(fmt.Sprintf("row type '%s' registered twice", name))
		}
		registry[name] = newRowType
	}
}

// Lookup returns the configured row types.
func Lookup(rowConfigs []RowConfig) ([]RowType, error) {
	rowTypes := make([]RowType, 0, len(rowConfigs))
	for _, rowConfig := range rowConfigs {
		newRowType, ok := registry[rowConfig.Name]
		if !ok {
			return nil, fmt.Errorf("unknown row type '%s' - available ones are: %s", rowConfig.Name, strings.Join(Names(), ", "))
		}
		rowType, err := newRowType(rowConfig.Settings)
		if err != nil {
			return nil, fmt.Errorf("invalid settings for row type '%s': %w", rowConfig.Name, err)
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

// SingleRow adapts a function generating exactly one row, not using the context, to Register's generate.
func SingleRow[S any](generate func(settings S) Row) func(ctx context.Context, settings S) []Row {
	return func(_ context.Context, settings S) []Row {
		return []Row{generate(settings)}
	}
}
