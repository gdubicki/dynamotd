package plugins

import (
	"context"
	"errors"
	"testing"

	"github.com/gdubicki/dynamotd/dynamotd"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/stretchr/testify/assert"
)

func mockDisks(t *testing.T, partitions []disk.PartitionStat, usages map[string]disk.UsageStat) {
	origPartitions, origUsage := getPartitions, getUsage
	t.Cleanup(func() { getPartitions, getUsage = origPartitions, origUsage })

	getPartitions = func(context.Context, bool) ([]disk.PartitionStat, error) { return partitions, nil }
	getUsage = func(_ context.Context, path string) (*disk.UsageStat, error) {
		usage, ok := usages[path]
		if !ok {
			return nil, errors.New("no such path")
		}
		usage.Path = path
		return &usage, nil
	}
}

var defaultDiskSpaceSettings = diskSpaceSettings{
	Thresholds: dynamotd.Thresholds{Warning: 75, Critical: 90},
	Inodes:     dynamotd.Thresholds{Warning: 75, Critical: 90},
	Always:     []string{"/"},
	Others:     true,
}

func labels(rows []dynamotd.Row) []string {
	var result []string
	for _, row := range rows {
		result = append(result, row.Label.String())
	}
	return result
}

func TestDiskSpaceShowsOnlyFilesystemsNeedingAttention(t *testing.T) {
	mockDisks(t, []disk.PartitionStat{
		{Device: "/dev/sda1", Mountpoint: "/", Fstype: "ext4"},
		{Device: "/dev/sda2", Mountpoint: "/home", Fstype: "ext4"},
		{Device: "/dev/sdb1", Mountpoint: "/var", Fstype: "xfs"},
		// bind mount of /var
		{Device: "/dev/sdb1", Mountpoint: "/var/lib/docker/something", Fstype: "xfs"},
		{Device: "/dev/sdc1", Mountpoint: "/data", Fstype: "ext4"},
		{Device: "/dev/loop0", Mountpoint: "/snap/core/1", Fstype: "squashfs"},
		{Device: "/dev/sdd1", Mountpoint: "/mnt/ro", Fstype: "ext4", Opts: []string{"ro"}},
		// macOS
		{Device: "devfs", Mountpoint: "/dev", Fstype: "devfs"},
	}, map[string]disk.UsageStat{
		"/":                         {Total: 100, Used: 10, UsedPercent: 10, InodesTotal: 100, InodesUsedPercent: 10},
		"/home":                     {Total: 100, Used: 50, UsedPercent: 50, InodesTotal: 100, InodesUsedPercent: 10},
		"/var":                      {Total: 100, Used: 95, UsedPercent: 95, InodesTotal: 100, InodesUsedPercent: 10},
		"/var/lib/docker/something": {Total: 100, Used: 95, UsedPercent: 95, InodesTotal: 100, InodesUsedPercent: 10},
		"/data":                     {Total: 100, Used: 10, UsedPercent: 10, InodesTotal: 100, InodesUsedPercent: 99},
		"/snap/core/1":              {Total: 100, Used: 100, UsedPercent: 100},
		"/mnt/ro":                   {Total: 100, Used: 100, UsedPercent: 100, InodesTotal: 100, InodesUsedPercent: 100},
		"/dev":                      {Total: 100, Used: 100, UsedPercent: 100, InodesTotal: 100, InodesUsedPercent: 100},
	})

	rows := DiskSpace(context.Background(), defaultDiskSpaceSettings)

	assert.Equal(t, []string{"Disk space (/)", "Disk space (/var)", "Disk space (/data)"}, labels(rows))
	assert.Equal(t, "10 B of 100 B disk space used (10%)", rows[0].Value.String())
	assert.Equal(t, "10 B of 100 B disk space used (10%), 99% inodes used", rows[2].Value.String())
	// the used space of /var is critical
	assert.Equal(t, dynamotd.ValueCriticalColor, rows[1].Value.Parts()[0].Color)
}

func TestDiskSpaceAlwaysShownAndOthersDisabled(t *testing.T) {
	mockDisks(t, []disk.PartitionStat{
		{Device: "/dev/sda1", Mountpoint: "/", Fstype: "ext4"},
		{Device: "/dev/sdb1", Mountpoint: "/var", Fstype: "ext4"},
	}, map[string]disk.UsageStat{
		"/":    {Total: 100, Used: 80, UsedPercent: 80},
		"/var": {Total: 100, Used: 99, UsedPercent: 99},
	})

	settings := defaultDiskSpaceSettings
	settings.Always = []string{"/", "/missing"}
	settings.Others = false

	rows := DiskSpace(context.Background(), settings)

	assert.Equal(t, []string{"Disk space (/)", "Disk space (/missing)"}, labels(rows))
	assert.Equal(t, dynamotd.ValueWarningColor, rows[0].Value.Parts()[0].Color)
	// no inodes info for filesystems without fixed number of inodes, like btrfs
	assert.Equal(t, "80 B of 100 B disk space used (80%)", rows[0].Value.String())
	assert.Equal(t, "can't get: error getting disk info: no such path", rows[1].Value.String())
}

func TestDiskSpaceSettingsValidation(t *testing.T) {
	settings := defaultDiskSpaceSettings
	settings.Inodes = dynamotd.Thresholds{Warning: 95, Critical: 90}
	assert.ErrorContains(t, settings.Validate(), "inodes: warning threshold")
}
