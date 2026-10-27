package plugins

import (
	"context"
	"fmt"
	"slices"

	"github.com/Tonyfilla/go-humanize"
	"github.com/gdubicki/dynamotd/dynamotd"
	"github.com/shirou/gopsutil/v4/disk"
)

type diskSpaceSettings struct {
	// thresholds for the disk space used, in %
	dynamotd.Thresholds `yaml:",inline"`
	// thresholds for the inodes used, in %
	Inodes dynamotd.Thresholds `yaml:"inodes"`
	// the paths of the filesystems that are always shown
	Always []string `yaml:"always"`
	// whether to also show the other filesystems that reached the warning threshold
	Others bool `yaml:"others"`
}

func (settings diskSpaceSettings) Validate() error {
	if err := settings.Thresholds.Validate(); err != nil {
		return err
	}
	if err := settings.Inodes.Validate(); err != nil {
		return fmt.Errorf("inodes: %w", err)
	}
	return nil
}

func init() {
	dynamotd.Register("Disk space", diskSpaceSettings{
		Thresholds: dynamotd.Thresholds{Warning: 75, Critical: 90},
		Inodes:     dynamotd.Thresholds{Warning: 75, Critical: 90},
		Always:     []string{"/"},
		Others:     true,
	}, DiskSpace, "diskspace")
}

// the types of filesystems that are always 100% full by design, incl. macOS' virtual ones
var fullByDesignFstypes = []string{"squashfs", "iso9660", "udf", "devfs", "autofs"}

// for mocking in tests
var getPartitions = disk.PartitionsWithContext
var getUsage = disk.UsageWithContext

// DiskSpace shows the disk space usage of the filesystems that are always shown, and of the other ones
// (excluding the read-only ones) that reached the warning threshold for space or inodes.
func DiskSpace(ctx context.Context, settings diskSpaceSettings) []dynamotd.Row {
	var rows []dynamotd.Row

	for _, path := range settings.Always {
		usage, err := getUsage(ctx, path)
		if err != nil {
			rows = append(rows, dynamotd.ErrorRow(diskSpaceLabel(path), fmt.Errorf("error getting disk info: %w", err)))
			continue
		}
		rows = append(rows, diskSpaceRow(usage, settings))
	}

	if !settings.Others {
		return rows
	}

	partitions, err := getPartitions(ctx, false)
	if err != nil {
		return append(rows, dynamotd.ErrorRow("Disk space", fmt.Errorf("error getting filesystems: %w", err)))
	}

	// a device can be mounted many times (bind mounts), so show each one only once, under its shortest path
	slices.SortStableFunc(partitions, func(a, b disk.PartitionStat) int { return len(a.Mountpoint) - len(b.Mountpoint) })
	shownDevices := map[string]bool{}
	for _, partition := range partitions {
		if slices.Contains(settings.Always, partition.Mountpoint) {
			shownDevices[partition.Device] = true
		}
	}

	for _, partition := range partitions {
		if shownDevices[partition.Device] || slices.Contains(fullByDesignFstypes, partition.Fstype) ||
			slices.Contains(partition.Opts, "ro") {
			continue
		}
		usage, err := getUsage(ctx, partition.Mountpoint)
		if err != nil {
			continue
		}
		if usage.UsedPercent >= settings.Warning || inodesUsedPercent(usage) >= settings.Inodes.Warning {
			rows = append(rows, diskSpaceRow(usage, settings))
			shownDevices[partition.Device] = true
		}
	}

	return rows
}

func diskSpaceLabel(path string) string {
	return fmt.Sprintf("Disk space (%s)", path)
}

// some filesystems, like btrfs, don't have a fixed number of inodes
func inodesUsedPercent(usage *disk.UsageStat) float64 {
	if usage.InodesTotal == 0 {
		return 0
	}
	return usage.InodesUsedPercent
}

func diskSpaceRow(usage *disk.UsageStat, settings diskSpaceSettings) dynamotd.Row {
	color := settings.Color(usage.UsedPercent)

	value := []dynamotd.ColorString{
		{Color: color, Text: humanize.IBytesCustomCeil(usage.Used, 2)},
		dynamotd.ValueDescription(" of "),
		dynamotd.ValueNeutral(humanize.IBytesCustomCeil(usage.Total, 0)),
		dynamotd.ValueDescription(" disk space used ("),
		{Color: color, Text: fmt.Sprintf("%0.0f%%", usage.UsedPercent)},
		dynamotd.ValueDescription(")"),
	}

	// show the inodes only if they are a problem, as they rarely are
	if inodesPercent := inodesUsedPercent(usage); inodesPercent >= settings.Inodes.Warning {
		value = append(value,
			dynamotd.ValueDescription(", "),
			dynamotd.ColorString{Color: settings.Inodes.Color(inodesPercent), Text: fmt.Sprintf("%0.0f%%", inodesPercent)},
			dynamotd.ValueDescription(" inodes used"),
		)
	}

	return dynamotd.Row{
		Label: dynamotd.SingleColorLabel(diskSpaceLabel(usage.Path)),
		Value: dynamotd.ToColorText(value...),
	}
}
