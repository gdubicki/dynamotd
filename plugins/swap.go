package plugins

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/Tonyfilla/go-humanize"
	"github.com/gdubicki/dynamotd/dynamotd"
	"github.com/shirou/gopsutil/v4/mem"
)

func init() {
	dynamotd.Register(swapLabel, dynamotd.Thresholds{Warning: 50, Critical: 80}, Swap, "swap")
}

const swapLabel = "Swap"

// for mocking in tests
var getSwap = mem.SwapMemoryWithContext
var vmstatPath = "/proc/vmstat"

// Swap shows the swap usage, with the thresholds in %, and the number of processes killed
// by the OOM killer since boot, if there were any.
func Swap(ctx context.Context, thresholds dynamotd.Thresholds) []dynamotd.Row {
	swap, err := getSwap(ctx)
	if err != nil {
		return []dynamotd.Row{dynamotd.ErrorRow(swapLabel, fmt.Errorf("error getting swap info: %w", err))}
	}

	var value []dynamotd.ColorString
	if swap.Total == 0 {
		value = append(value, dynamotd.ValueDescription("no swap"))
	} else {
		color := thresholds.Color(swap.UsedPercent)
		value = append(value,
			dynamotd.ColorString{Color: color, Text: humanize.IBytesCustomCeil(swap.Used, 2)},
			dynamotd.ValueDescription(" of "),
			dynamotd.ValueNeutral(humanize.IBytesCustomCeil(swap.Total, 2)),
			dynamotd.ValueDescription(" swap used ("),
			dynamotd.ColorString{Color: color, Text: fmt.Sprintf("%0.0f%%", swap.UsedPercent)},
			dynamotd.ValueDescription(")"),
		)
	}

	// it's not available on macOS and old Linux kernels
	if oomKills, err := getOOMKills(); err == nil && oomKills > 0 {
		unit := "OOM kill"
		if oomKills > 1 {
			unit += "s"
		}
		value = append(value,
			dynamotd.ValueDescription(", "),
			dynamotd.ValueCritical(fmt.Sprintf("%d %s", oomKills, unit)),
			dynamotd.ValueDescription(" since boot"),
		)
	}

	return []dynamotd.Row{{
		Label: dynamotd.SingleColorLabel(swapLabel),
		Value: dynamotd.ToColorText(value...),
	}}
}

// getOOMKills returns the number of processes killed by the OOM killer since boot.
func getOOMKills() (uint64, error) {
	file, err := os.Open(vmstatPath)
	if err != nil {
		return 0, err
	}
	defer func() { _ = file.Close() }()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if value, found := strings.CutPrefix(scanner.Text(), "oom_kill "); found {
			return strconv.ParseUint(value, 10, 64)
		}
	}
	if err := scanner.Err(); err != nil {
		return 0, err
	}
	return 0, fmt.Errorf("no oom_kill in %s", vmstatPath)
}
